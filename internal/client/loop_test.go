package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

func newTestMockServer() *mockserver.Server {
	return &mockserver.Server{
		Rules: []mockserver.Rule{{Path: "/paid", Amount: 100, Asset: "USDC", Recipient: "0xsandbox"}},
		Next: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("paid content"))
		}),
	}
}

func TestLoopDoPaysChallenge(t *testing.T) {
	srv := httptest.NewServer(newTestMockServer())
	defer srv.Close()

	loop := &Loop{Signer: FakeSigner{}}
	result, err := loop.Do(context.Background(), http.MethodGet, srv.URL+"/paid")
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if result.FinalStatusCode != http.StatusOK {
		t.Fatalf("FinalStatusCode = %d, want %d", result.FinalStatusCode, http.StatusOK)
	}
	if !result.Paid {
		t.Fatal("Paid = false, want true")
	}
	if string(result.FinalBody) != "paid content" {
		t.Fatalf("FinalBody = %q, want %q", result.FinalBody, "paid content")
	}
	if len(result.Steps) != 3 {
		t.Fatalf("len(Steps) = %d, want 3: %+v", len(result.Steps), result.Steps)
	}
	if result.Steps[1].Descriptor == nil {
		t.Fatal("Steps[1].Descriptor is nil, want the received challenge")
	}
	if result.Steps[2].Proof == nil {
		t.Fatal("Steps[2].Proof is nil, want the presented proof")
	}
}

func TestLoopDoSkipsUnchallengedRoute(t *testing.T) {
	srv := httptest.NewServer(newTestMockServer())
	defer srv.Close()

	loop := &Loop{Signer: FakeSigner{}}
	result, err := loop.Do(context.Background(), http.MethodGet, srv.URL+"/free")
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if result.Paid {
		t.Fatal("Paid = true, want false for an unprotected route")
	}
	if len(result.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1: %+v", len(result.Steps), result.Steps)
	}
}

type erroringSigner struct{}

func (erroringSigner) Sign(paywall.Descriptor) (paywall.Proof, error) {
	return paywall.Proof{}, errors.New("signer: refused to sign")
}

func TestLoopDoReturnsSignerError(t *testing.T) {
	srv := httptest.NewServer(newTestMockServer())
	defer srv.Close()

	loop := &Loop{Signer: erroringSigner{}}
	if _, err := loop.Do(context.Background(), http.MethodGet, srv.URL+"/paid"); err == nil {
		t.Fatal("Do() error = nil, want signer error surfaced")
	}
}
