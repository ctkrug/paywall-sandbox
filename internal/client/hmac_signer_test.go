package client

import (
	"testing"
	"time"
)

func TestHMACSignerSignsValidDescriptor(t *testing.T) {
	d := validDescriptor()
	proof, err := HMACSigner{Key: []byte("shared-secret")}.Sign(d)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if proof.Nonce != d.Nonce {
		t.Fatalf("proof.Nonce = %q, want %q", proof.Nonce, d.Nonce)
	}
	if proof.Scheme != HMACScheme {
		t.Fatalf("proof.Scheme = %q, want %q", proof.Scheme, HMACScheme)
	}
	if proof.Signature == "" {
		t.Fatal("proof.Signature is empty")
	}
}

func TestHMACSignerIsDeterministic(t *testing.T) {
	d := validDescriptor()
	key := []byte("shared-secret")

	first, err := HMACSigner{Key: key}.Sign(d)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	second, err := HMACSigner{Key: key}.Sign(d)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if first.Signature != second.Signature {
		t.Fatalf("signatures differ for same key/nonce: %q vs %q", first.Signature, second.Signature)
	}
}

func TestHMACSignerRejectsExpiredDescriptor(t *testing.T) {
	d := validDescriptor()
	d.ExpiresAt = time.Now().Add(-time.Second)

	if _, err := (HMACSigner{Key: []byte("k")}).Sign(d); err == nil {
		t.Fatal("Sign() error = nil, want error for expired descriptor")
	}
}

func TestHMACSignerRejectsInvalidDescriptor(t *testing.T) {
	d := validDescriptor()
	d.Recipient = ""

	if _, err := (HMACSigner{Key: []byte("k")}).Sign(d); err == nil {
		t.Fatal("Sign() error = nil, want error for invalid descriptor")
	}
}
