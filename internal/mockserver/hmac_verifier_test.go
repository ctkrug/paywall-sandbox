package mockserver

import (
	"testing"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

func TestHMACVerifierAcceptsCorrectSignature(t *testing.T) {
	key := []byte("shared-secret")
	desc := paywall.Descriptor{
		Amount: 100, Asset: "USDC", Recipient: "0xsandbox",
		Nonce: "nonce-1", ExpiresAt: time.Now().Add(time.Minute),
	}
	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: HMACScheme, Signature: hmacSign(key, desc.Nonce)}

	if err := (HMACVerifier{Key: key}).Verify(desc, proof); err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
}

func TestHMACVerifierRejectsWrongSignature(t *testing.T) {
	desc := paywall.Descriptor{Nonce: "nonce-1"}
	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: HMACScheme, Signature: "not-the-right-signature"}

	if err := (HMACVerifier{Key: []byte("shared-secret")}).Verify(desc, proof); err == nil {
		t.Fatal("Verify() error = nil, want mismatch error")
	}
}

func TestHMACVerifierRejectsWrongKey(t *testing.T) {
	desc := paywall.Descriptor{Nonce: "nonce-1"}
	proof := paywall.Proof{Nonce: desc.Nonce, Scheme: HMACScheme, Signature: hmacSign([]byte("secret-a"), desc.Nonce)}

	if err := (HMACVerifier{Key: []byte("secret-b")}).Verify(desc, proof); err == nil {
		t.Fatal("Verify() error = nil, want mismatch error for wrong key")
	}
}
