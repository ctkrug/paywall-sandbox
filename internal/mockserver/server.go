package mockserver

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

// FakeScheme is the only proof scheme this sandbox accepts. It exists to
// exercise the challenge/response shape end to end and settles nothing for
// real — see docs/PROTOCOL.md.
const FakeScheme = "fake"

const defaultTTL = 5 * time.Minute

// Server issues HTTP 402 challenges for requests matching one of its Rules
// and forwards to Next once a valid, unexpired, unused proof is presented.
type Server struct {
	Rules []Rule
	Next  http.Handler
	// TTL is how long an issued challenge remains valid. Zero means
	// defaultTTL; a negative value is honored as-is, which is useful for
	// deterministically testing expiry.
	TTL time.Duration

	mu     sync.Mutex
	issued map[string]time.Time // nonce -> expiry
}

func (s *Server) ttl() time.Duration {
	if s.TTL == 0 {
		return defaultTTL
	}
	return s.TTL
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rule, ok := s.matchRule(req)
	if !ok {
		s.next().ServeHTTP(w, req)
		return
	}

	if s.acceptProof(req) {
		s.next().ServeHTTP(w, req)
		return
	}

	s.challenge(w, rule)
}

func (s *Server) next() http.Handler {
	if s.Next != nil {
		return s.Next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) matchRule(req *http.Request) (Rule, bool) {
	for _, r := range s.Rules {
		if r.Matches(req) {
			return r, true
		}
	}
	return Rule{}, false
}

func (s *Server) challenge(w http.ResponseWriter, rule Rule) {
	desc := paywall.Descriptor{
		Amount:    rule.Amount,
		Asset:     rule.Asset,
		Recipient: rule.Recipient,
		Nonce:     s.issueNonce(),
		ExpiresAt: time.Now().Add(s.ttl()),
	}

	body, err := desc.Encode()
	if err != nil {
		http.Error(w, "failed to encode payment descriptor", http.StatusInternalServerError)
		return
	}

	w.Header().Set(paywall.HeaderChallenge, string(body))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	_, _ = w.Write(body)
}

func (s *Server) issueNonce() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	nonce := hex.EncodeToString(buf)

	s.mu.Lock()
	if s.issued == nil {
		s.issued = make(map[string]time.Time)
	}
	s.issued[nonce] = time.Now().Add(s.ttl())
	s.mu.Unlock()

	return nonce
}

// acceptProof reports whether req carries a valid, unexpired proof for a
// nonce this Server issued, consuming it in the process so a proof can't be
// replayed against a second request.
func (s *Server) acceptProof(req *http.Request) bool {
	raw := req.Header.Get(paywall.HeaderProof)
	if raw == "" {
		return false
	}

	proof, err := paywall.DecodeProof([]byte(raw))
	if err != nil || proof.Scheme != FakeScheme {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expiry, ok := s.issued[proof.Nonce]
	if !ok || time.Now().After(expiry) {
		return false
	}

	delete(s.issued, proof.Nonce)
	return true
}
