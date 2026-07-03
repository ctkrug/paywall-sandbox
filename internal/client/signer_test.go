package client

import (
	"testing"
	"time"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

func validDescriptor() paywall.Descriptor {
	return paywall.Descriptor{
		Amount:    100,
		Asset:     "USDC",
		Recipient: "0xsandbox",
		Nonce:     "nonce-1",
		ExpiresAt: time.Now().Add(time.Minute),
	}
}

func TestFakeSignerSignsValidDescriptor(t *testing.T) {
	proof, err := FakeSigner{}.Sign(validDescriptor())
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if proof.Nonce != "nonce-1" {
		t.Fatalf("proof.Nonce = %q, want %q", proof.Nonce, "nonce-1")
	}
	if proof.Scheme != FakeScheme {
		t.Fatalf("proof.Scheme = %q, want %q", proof.Scheme, FakeScheme)
	}
	if proof.Signature == "" {
		t.Fatal("proof.Signature is empty")
	}
}

func TestFakeSignerRejectsExpiredDescriptor(t *testing.T) {
	d := validDescriptor()
	d.ExpiresAt = time.Now().Add(-time.Second)

	if _, err := (FakeSigner{}).Sign(d); err == nil {
		t.Fatal("Sign() error = nil, want error for expired descriptor")
	}
}

func TestFakeSignerRejectsInvalidDescriptor(t *testing.T) {
	d := validDescriptor()
	d.Amount = 0

	if _, err := (FakeSigner{}).Sign(d); err == nil {
		t.Fatal("Sign() error = nil, want error for invalid descriptor")
	}
}
