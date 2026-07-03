package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

// HMACScheme is the scheme HMACSigner produces, matching
// mockserver.HMACScheme. See docs/PROTOCOL.md.
const HMACScheme = "hmac-sha256"

// HMACSigner settles a Descriptor by HMAC-signing its nonce with a shared
// key. It exists alongside FakeSigner as a second, still-offline settlement
// simulation — see docs/VISION.md.
type HMACSigner struct {
	Key []byte
}

// Sign implements Signer.
func (s HMACSigner) Sign(d paywall.Descriptor) (paywall.Proof, error) {
	if err := d.Validate(); err != nil {
		return paywall.Proof{}, fmt.Errorf("client: invalid descriptor: %w", err)
	}
	if d.Expired(time.Now()) {
		return paywall.Proof{}, fmt.Errorf("client: descriptor for nonce %s expired at %s", d.Nonce, d.ExpiresAt)
	}

	mac := hmac.New(sha256.New, s.Key)
	mac.Write([]byte(d.Nonce))

	return paywall.Proof{
		Nonce:     d.Nonce,
		Scheme:    HMACScheme,
		Signature: hex.EncodeToString(mac.Sum(nil)),
	}, nil
}
