package scenario

import (
	"net/http/httptest"

	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

// newServer starts an in-process mock server configured from s: its rule
// set, plus a Verifier for every scheme a step in s might use. The caller
// must Close the returned server when done.
func (s Scenario) newServer() (*httptest.Server, error) {
	rules, err := s.rules()
	if err != nil {
		return nil, err
	}

	srv := &mockserver.Server{Rules: rules, Verifiers: s.verifiers()}
	return httptest.NewServer(srv), nil
}

// verifiers builds the Verifier set the mock server accepts: FakeScheme
// always, plus HMACScheme whenever an HMACKey is configured.
func (s Scenario) verifiers() map[string]mockserver.Verifier {
	verifiers := map[string]mockserver.Verifier{
		mockserver.FakeScheme: mockserver.VerifierFunc(func(paywall.Descriptor, paywall.Proof) error { return nil }),
	}
	if s.HMACKey != "" {
		verifiers[mockserver.HMACScheme] = mockserver.HMACVerifier{Key: []byte(s.HMACKey)}
	}
	return verifiers
}
