package scenario

import (
	"fmt"

	"github.com/ctkrug/paywall-sandbox/internal/client"
)

// signerFor resolves the client.Signer a step's Scheme names. An empty
// Scheme defaults to client.FakeScheme.
func signerFor(scheme string, hmacKey []byte) (client.Signer, error) {
	switch scheme {
	case "", client.FakeScheme:
		return client.FakeSigner{}, nil
	case client.HMACScheme:
		return client.HMACSigner{Key: hmacKey}, nil
	default:
		return nil, fmt.Errorf("scenario: unknown scheme %q", scheme)
	}
}
