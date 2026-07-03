package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ctkrug/paywall-sandbox/internal/paywall"
)

// Step records one request/response exchange in a Loop run, for inspection
// (e.g. --verbose CLI output).
type Step struct {
	Label      string
	Method     string
	URL        string
	StatusCode int
	Descriptor *paywall.Descriptor
	Proof      *paywall.Proof
}

// Result is the outcome of a Loop run: the final response plus a trace of
// every step taken to get there.
type Result struct {
	// FinalStatusCode and FinalBody describe the last response received,
	// whether or not a challenge was ever issued.
	FinalStatusCode int
	FinalBody       []byte
	// Paid reports whether a 402 challenge was settled and retried.
	Paid bool
	// Steps traces every request sent, in order.
	Steps []Step
}

// Loop drives the challenge -> pay -> retry exchange described in
// docs/PROTOCOL.md against any target, mock or real.
type Loop struct {
	// HTTPClient sends requests. A nil value uses http.DefaultClient.
	HTTPClient *http.Client
	// Signer builds a Proof from a received Descriptor. Required.
	Signer Signer
}

func (l *Loop) httpClient() *http.Client {
	if l.HTTPClient != nil {
		return l.HTTPClient
	}
	return http.DefaultClient
}

// Do sends method/url and, if challenged with a 402, settles it via Signer
// and retries once. It returns the final response and a step-by-step trace
// regardless of whether payment was required.
func (l *Loop) Do(ctx context.Context, method, url string) (*Result, error) {
	result := &Result{}

	status, body, resp, err := l.send(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("client: initial request: %w", err)
	}
	result.Steps = append(result.Steps, Step{Label: "initial request", Method: method, URL: url, StatusCode: status})

	if status != http.StatusPaymentRequired {
		result.FinalStatusCode = status
		result.FinalBody = body
		return result, nil
	}

	desc, err := descriptorFromResponse(resp, body)
	if err != nil {
		return nil, fmt.Errorf("client: parsing payment descriptor: %w", err)
	}
	result.Steps = append(result.Steps, Step{
		Label: "402 challenge received", Method: method, URL: url, StatusCode: status, Descriptor: &desc,
	})

	proof, err := l.Signer.Sign(desc)
	if err != nil {
		return nil, fmt.Errorf("client: signing proof: %w", err)
	}

	encodedProof, err := proof.Encode()
	if err != nil {
		return nil, fmt.Errorf("client: encoding proof: %w", err)
	}

	status, body, _, err = l.send(ctx, method, url, map[string]string{paywall.HeaderProof: string(encodedProof)})
	if err != nil {
		return nil, fmt.Errorf("client: retry with proof: %w", err)
	}
	result.Paid = true
	result.Steps = append(result.Steps, Step{
		Label: "retry with proof", Method: method, URL: url, StatusCode: status, Proof: &proof,
	})
	result.FinalStatusCode = status
	result.FinalBody = body

	return result, nil
}

func (l *Loop) send(ctx context.Context, method, url string, headers map[string]string) (int, []byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := l.httpClient().Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return resp.StatusCode, body, resp, nil
}

func descriptorFromResponse(resp *http.Response, body []byte) (paywall.Descriptor, error) {
	if header := resp.Header.Get(paywall.HeaderChallenge); header != "" {
		return paywall.DecodeDescriptor([]byte(header))
	}
	return paywall.DecodeDescriptor(body)
}
