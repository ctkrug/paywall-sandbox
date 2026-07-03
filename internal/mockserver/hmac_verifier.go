package mockserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

// HMACScheme is a second, still-offline settlement simulation: Proof.Signature
// must be hex(HMAC-SHA256(Key, Descriptor.Nonce)). It proves a client held
// the same shared secret as the server, one step closer to real settlement
// evidence than FakeScheme's unconditional accept — see docs/VISION.md.
const HMACScheme = "hmac-sha256"

// HMACVerifier verifies HMACScheme proofs against a shared Key.
type HMACVerifier struct {
	Key []byte
}

// Verify implements Verifier.
func (v HMACVerifier) Verify(d paywall.Descriptor, p paywall.Proof) error {
	want := hmacSign(v.Key, d.Nonce)
	if !hmac.Equal([]byte(want), []byte(p.Signature)) {
		return fmt.Errorf("mockserver: hmac signature mismatch for nonce %s", d.Nonce)
	}
	return nil
}

func hmacSign(key []byte, nonce string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(nonce))
	return hex.EncodeToString(mac.Sum(nil))
}
