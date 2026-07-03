package paywall

import (
	"testing"
	"time"
)

func validDescriptor() Descriptor {
	return Descriptor{
		Amount:    100,
		Asset:     "USDC",
		Recipient: "0xabc123",
		Nonce:     "nonce-1",
		ExpiresAt: time.Now().Add(time.Minute),
	}
}

func TestDescriptorValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(d Descriptor) Descriptor
		wantErr bool
	}{
		{"valid", func(d Descriptor) Descriptor { return d }, false},
		{"zero amount", func(d Descriptor) Descriptor { d.Amount = 0; return d }, true},
		{"empty asset", func(d Descriptor) Descriptor { d.Asset = ""; return d }, true},
		{"empty recipient", func(d Descriptor) Descriptor { d.Recipient = ""; return d }, true},
		{"empty nonce", func(d Descriptor) Descriptor { d.Nonce = ""; return d }, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.mutate(validDescriptor()).Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestDescriptorExpired(t *testing.T) {
	d := validDescriptor()
	d.ExpiresAt = time.Now().Add(-time.Second)
	if !d.Expired(time.Now()) {
		t.Fatal("expected descriptor to be expired")
	}

	d.ExpiresAt = time.Now().Add(time.Minute)
	if d.Expired(time.Now()) {
		t.Fatal("expected descriptor to not be expired")
	}
}

func TestDescriptorRoundTrip(t *testing.T) {
	want := validDescriptor()
	data, err := want.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	got, err := DecodeDescriptor(data)
	if err != nil {
		t.Fatalf("DecodeDescriptor() error = %v", err)
	}
	if got.Amount != want.Amount || got.Asset != want.Asset || got.Nonce != want.Nonce {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestProofRoundTrip(t *testing.T) {
	want := Proof{Nonce: "nonce-1", Scheme: "fake", Signature: "sig"}
	data, err := want.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	got, err := DecodeProof(data)
	if err != nil {
		t.Fatalf("DecodeProof() error = %v", err)
	}
	if got != want {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, want)
	}
}
