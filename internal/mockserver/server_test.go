package mockserver

import (
	"errors"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

func newTestServer(ttl time.Duration) *Server {
	return &Server{
		Rules: []Rule{{Path: "/paid", Amount: 100, Asset: "USDC", Recipient: "0xsandbox"}},
		TTL:   ttl,
	}
}

func TestServerChallengesUnpaidRequest(t *testing.T) {
	s := newTestServer(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/paid", nil)

	s.ServeHTTP(rec, req)

	if rec.Code != 402 {
		t.Fatalf("status = %d, want 402", rec.Code)
	}
	desc, err := paywall.DecodeDescriptor(rec.Body.Bytes())
	if err != nil {
		t.Fatalf("DecodeDescriptor() error = %v", err)
	}
	if desc.Amount != 100 || desc.Asset != "USDC" || desc.Nonce == "" {
		t.Fatalf("unexpected descriptor: %+v", desc)
	}
}

func TestServerAcceptsValidProof(t *testing.T) {
	s := newTestServer(0)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	desc, _ := paywall.DecodeDescriptor(rec.Body.Bytes())

	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: FakeScheme, Signature: "n/a"}
	proofBody, _ := proof.Encode()

	req := httptest.NewRequest("GET", "/paid", nil)
	req.Header.Set(paywall.HeaderProof, string(proofBody))

	rec2 := httptest.NewRecorder()
	s.ServeHTTP(rec2, req)

	if rec2.Code != 200 {
		t.Fatalf("status = %d, want 200", rec2.Code)
	}
}

func TestServerRejectsReplayedProof(t *testing.T) {
	s := newTestServer(0)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	desc, _ := paywall.DecodeDescriptor(rec.Body.Bytes())

	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: FakeScheme, Signature: "n/a"}
	proofBody, _ := proof.Encode()

	req := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/paid", nil)
		r.Header.Set(paywall.HeaderProof, string(proofBody))
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, r)
		return rec
	}

	first := req()
	if first.Code != 200 {
		t.Fatalf("first attempt status = %d, want 200", first.Code)
	}

	second := req()
	if second.Code != 402 {
		t.Fatalf("replayed attempt status = %d, want 402", second.Code)
	}
}

func TestServerRejectsExpiredProof(t *testing.T) {
	s := newTestServer(-time.Second)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	desc, _ := paywall.DecodeDescriptor(rec.Body.Bytes())

	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: FakeScheme, Signature: "n/a"}
	proofBody, _ := proof.Encode()

	req := httptest.NewRequest("GET", "/paid", nil)
	req.Header.Set(paywall.HeaderProof, string(proofBody))

	rec2 := httptest.NewRecorder()
	s.ServeHTTP(rec2, req)

	if rec2.Code != 402 {
		t.Fatalf("status = %d, want 402 for expired nonce", rec2.Code)
	}
}

func TestServerUsesCustomVerifier(t *testing.T) {
	const customScheme = "custom"
	const wantSignature = "the-right-signature"

	s := newTestServer(0)
	s.Verifiers = map[string]Verifier{
		customScheme: VerifierFunc(func(_ paywall.Descriptor, p paywall.Proof) error {
			if p.Signature != wantSignature {
				return errors.New("signature mismatch")
			}
			return nil
		}),
	}

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	desc, _ := paywall.DecodeDescriptor(rec.Body.Bytes())

	wrongProof := paywall.Proof{Nonce: desc.Nonce, Scheme: customScheme, Signature: "wrong"}
	wrongBody, _ := wrongProof.Encode()
	wrongReq := httptest.NewRequest("GET", "/paid", nil)
	wrongReq.Header.Set(paywall.HeaderProof, string(wrongBody))
	wrongRec := httptest.NewRecorder()
	s.ServeHTTP(wrongRec, wrongReq)
	if wrongRec.Code != 402 {
		t.Fatalf("wrong signature status = %d, want 402", wrongRec.Code)
	}

	rightProof := paywall.Proof{Nonce: desc.Nonce, Scheme: customScheme, Signature: wantSignature}
	rightBody, _ := rightProof.Encode()
	rightReq := httptest.NewRequest("GET", "/paid", nil)
	rightReq.Header.Set(paywall.HeaderProof, string(rightBody))
	rightRec := httptest.NewRecorder()
	s.ServeHTTP(rightRec, rightReq)
	if rightRec.Code != 200 {
		t.Fatalf("right signature status = %d, want 200", rightRec.Code)
	}
}

func TestServerRejectsUnregisteredScheme(t *testing.T) {
	s := newTestServer(0)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	desc, _ := paywall.DecodeDescriptor(rec.Body.Bytes())

	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: "not-registered", Signature: "n/a"}
	proofBody, _ := proof.Encode()

	req := httptest.NewRequest("GET", "/paid", nil)
	req.Header.Set(paywall.HeaderProof, string(proofBody))

	rec2 := httptest.NewRecorder()
	s.ServeHTTP(rec2, req)

	if rec2.Code != 402 {
		t.Fatalf("status = %d, want 402 for an unregistered scheme", rec2.Code)
	}
}

func TestServerPassesThroughUnmatchedRoutes(t *testing.T) {
	s := newTestServer(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/free", nil)

	s.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 for unmatched route", rec.Code)
	}
}

// TestServerSweepsExpiredNonces guards against unbounded memory growth on a
// long-running Server: challenges that expire without ever being retried
// must not accumulate in the issued map forever.
func TestServerSweepsExpiredNonces(t *testing.T) {
	s := newTestServer(-time.Second)

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	}

	s.mu.Lock()
	got := len(s.issued)
	s.mu.Unlock()
	if got != 1 {
		t.Fatalf("len(issued) = %d, want 1 (all prior entries were already expired)", got)
	}
}

// TestServerDeletesExpiredNonceOnAttemptedUse checks that presenting a
// proof for an already-expired nonce also reclaims that entry immediately,
// rather than waiting for the next challenge to sweep it.
func TestServerDeletesExpiredNonceOnAttemptedUse(t *testing.T) {
	s := newTestServer(-time.Second)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
	desc, _ := paywall.DecodeDescriptor(rec.Body.Bytes())

	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: FakeScheme, Signature: "n/a"}
	proofBody, _ := proof.Encode()
	req := httptest.NewRequest("GET", "/paid", nil)
	req.Header.Set(paywall.HeaderProof, string(proofBody))
	s.ServeHTTP(httptest.NewRecorder(), req)

	s.mu.Lock()
	_, stillPresent := s.issued[desc.Nonce]
	s.mu.Unlock()
	if stillPresent {
		t.Fatal("expired nonce was not reclaimed on attempted use")
	}
}

// TestServerHandlesConcurrentChallenges drives many independent
// challenge->pay round trips through the same Server at once, guarding
// the shared issued map (and its new sweep-on-store cleanup) against
// races and cross-talk between nonces. Run with -race to be meaningful.
func TestServerHandlesConcurrentChallenges(t *testing.T) {
	s := newTestServer(time.Minute)

	const workers = 50
	var wg sync.WaitGroup
	errs := make(chan string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, httptest.NewRequest("GET", "/paid", nil))
			if rec.Code != 402 {
				errs <- "challenge status != 402"
				return
			}
			desc, err := paywall.DecodeDescriptor(rec.Body.Bytes())
			if err != nil {
				errs <- "decode: " + err.Error()
				return
			}

			proof := paywall.Proof{Nonce: desc.Nonce, Scheme: FakeScheme, Signature: "n/a"}
			proofBody, _ := proof.Encode()
			req := httptest.NewRequest("GET", "/paid", nil)
			req.Header.Set(paywall.HeaderProof, string(proofBody))

			payRec := httptest.NewRecorder()
			s.ServeHTTP(payRec, req)
			if payRec.Code != 200 {
				errs <- "pay status != 200"
			}
		}()
	}

	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
