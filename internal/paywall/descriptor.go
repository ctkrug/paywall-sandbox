// Package paywall defines the shape of an HTTP 402 payment challenge and
// the proof a client presents to satisfy it.
package paywall

import (
	"encoding/json"
	"errors"
	"time"
)

// HeaderChallenge is the response header carrying the JSON-encoded Descriptor.
const HeaderChallenge = "X-Payment-Required"

// HeaderProof is the request header a client sets on retry to present proof
// of payment against a previously issued Descriptor.
const HeaderProof = "X-Payment"

// Descriptor describes what a client must pay to be granted access to a
// resource, and by when. It is served as the body (and mirrored into
// HeaderChallenge) of every 402 response.
type Descriptor struct {
	// Amount is the price in the smallest unit of Asset (e.g. cents, wei).
	Amount uint64 `json:"amount"`
	// Asset identifies the currency or token, e.g. "USD" or "USDC".
	Asset string `json:"asset"`
	// Recipient is the address or account payment must settle to.
	Recipient string `json:"recipient"`
	// Nonce uniquely identifies this challenge so a proof can't be replayed
	// against a different request.
	Nonce string `json:"nonce"`
	// ExpiresAt is when this challenge stops being valid.
	ExpiresAt time.Time `json:"expiresAt"`
}

// Expired reports whether the descriptor is no longer valid at t.
func (d Descriptor) Expired(t time.Time) bool {
	return t.After(d.ExpiresAt)
}

// Validate checks that the descriptor is internally consistent.
func (d Descriptor) Validate() error {
	if d.Amount == 0 {
		return errors.New("paywall: amount must be greater than zero")
	}
	if d.Asset == "" {
		return errors.New("paywall: asset must not be empty")
	}
	if d.Recipient == "" {
		return errors.New("paywall: recipient must not be empty")
	}
	if d.Nonce == "" {
		return errors.New("paywall: nonce must not be empty")
	}
	return nil
}

// Proof is what a client attaches (as JSON in HeaderProof) on retry to
// demonstrate the Descriptor's challenge has been settled.
type Proof struct {
	// Nonce must match the Descriptor.Nonce it satisfies.
	Nonce string `json:"nonce"`
	// Scheme names the settlement mechanism, e.g. "fake" or "x402-exact".
	Scheme string `json:"scheme"`
	// Signature is scheme-specific evidence of settlement.
	Signature string `json:"signature"`
}

// Encode marshals d to its wire form.
func (d Descriptor) Encode() ([]byte, error) {
	return json.Marshal(d)
}

// DecodeDescriptor parses a Descriptor from its wire form.
func DecodeDescriptor(data []byte) (Descriptor, error) {
	var d Descriptor
	if err := json.Unmarshal(data, &d); err != nil {
		return Descriptor{}, err
	}
	return d, nil
}

// Encode marshals p to its wire form.
func (p Proof) Encode() ([]byte, error) {
	return json.Marshal(p)
}

// DecodeProof parses a Proof from its wire form.
func DecodeProof(data []byte) (Proof, error) {
	var p Proof
	if err := json.Unmarshal(data, &p); err != nil {
		return Proof{}, err
	}
	return p, nil
}
