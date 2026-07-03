package mockserver

import "github.com/ctkrug/paywall-sandbox/internal/paywall"

// Verifier validates a Proof's Signature against the Descriptor it claims
// to satisfy. Nonce replay and expiry are checked by Server independently
// of scheme; Verify only judges whether the signature itself is valid for
// that scheme. Implementations are pluggable so schemes beyond FakeScheme
// can be registered without changing Server — see docs/BACKLOG.md.
type Verifier interface {
	Verify(paywall.Descriptor, paywall.Proof) error
}

// VerifierFunc adapts a function to a Verifier.
type VerifierFunc func(paywall.Descriptor, paywall.Proof) error

// Verify implements Verifier.
func (f VerifierFunc) Verify(d paywall.Descriptor, p paywall.Proof) error {
	return f(d, p)
}

// acceptAny is the Verifier for FakeScheme: it accepts any signature once
// the nonce has checked out, matching the sandbox's v1 behavior — see
// docs/PROTOCOL.md.
var acceptAny Verifier = VerifierFunc(func(paywall.Descriptor, paywall.Proof) error { return nil })

// defaultVerifiers is used when Server.Verifiers is nil, preserving the
// original fake-only behavior.
func defaultVerifiers() map[string]Verifier {
	return map[string]Verifier{FakeScheme: acceptAny}
}
