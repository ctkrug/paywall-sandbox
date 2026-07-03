package mockserver

import (
	"net/http/httptest"
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

func TestServerPassesThroughUnmatchedRoutes(t *testing.T) {
	s := newTestServer(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/free", nil)

	s.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 for unmatched route", rec.Code)
	}
}
