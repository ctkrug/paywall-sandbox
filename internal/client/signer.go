// Package client implements the client side of the 402 challenge/response
// exchange: sending a request, settling a received Descriptor into a Proof,
// and retrying. See docs/PROTOCOL.md for the wire format this drives.
package client

import (
	"fmt"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

// FakeScheme is the scheme FakeSigner produces, matching the only scheme
// mockserver.Server accepts. See docs/PROTOCOL.md.
const FakeScheme = "fake"

// Signer settles a Descriptor's challenge and builds the Proof a Loop
// presents on retry. Implementations are pluggable so schemes beyond
// FakeScheme can be added without changing Loop — see docs/BACKLOG.md.
type Signer interface {
	Sign(paywall.Descriptor) (paywall.Proof, error)
}

// FakeSigner settles a Descriptor by fabricating a signature. It exists to
// exercise the protocol shape end to end and proves nothing about real
// settlement — see docs/PROTOCOL.md.
type FakeSigner struct{}

// Sign implements Signer.
func (FakeSigner) Sign(d paywall.Descriptor) (paywall.Proof, error) {
	if err := d.Validate(); err != nil {
		return paywall.Proof{}, fmt.Errorf("client: invalid descriptor: %w", err)
	}
	if d.Expired(time.Now()) {
		return paywall.Proof{}, fmt.Errorf("client: descriptor for nonce %s expired at %s", d.Nonce, d.ExpiresAt)
	}
	return paywall.Proof{
		Nonce:     d.Nonce,
		Scheme:    FakeScheme,
		Signature: "fake-signature-" + d.Nonce,
	}, nil
}
