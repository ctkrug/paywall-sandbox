package main

import (
	"testing"

	"github.com/ctkrug/paywall-sandbox/internal/client"
)

func TestResolveSignerFake(t *testing.T) {
	signer, err := resolveSigner(client.FakeScheme, "")
	if err != nil {
		t.Fatalf("resolveSigner: %v", err)
	}
	if _, ok := signer.(client.FakeSigner); !ok {
		t.Errorf("resolveSigner(%q) = %T, want client.FakeSigner", client.FakeScheme, signer)
	}
}

func TestResolveSignerHMAC(t *testing.T) {
	signer, err := resolveSigner(client.HMACScheme, "shared-secret")
	if err != nil {
		t.Fatalf("resolveSigner: %v", err)
	}
	hmacSigner, ok := signer.(client.HMACSigner)
	if !ok {
		t.Fatalf("resolveSigner(%q) = %T, want client.HMACSigner", client.HMACScheme, signer)
	}
	if string(hmacSigner.Key) != "shared-secret" {
		t.Errorf("HMACSigner.Key = %q, want %q", hmacSigner.Key, "shared-secret")
	}
}

func TestResolveSignerHMACRequiresKey(t *testing.T) {
	if _, err := resolveSigner(client.HMACScheme, ""); err == nil {
		t.Fatal("resolveSigner: want error for hmac-sha256 with no key, got nil")
	}
}

func TestResolveSignerUnknown(t *testing.T) {
	if _, err := resolveSigner("bogus", ""); err == nil {
		t.Fatal("resolveSigner: want error for unknown scheme, got nil")
	}
}
