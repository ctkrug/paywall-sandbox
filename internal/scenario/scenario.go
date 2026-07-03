// Package scenario implements a declarative format for scripting an
// expected challenge/response sequence against an in-process mock server,
// and running it as a pass/fail assertion. See docs/PROTOCOL.md for the
// file format and the wire format it exercises.
package scenario

import "github.com/ctkrug/paywall-sandbox/internal/mockserver"

// Scenario is one runnable script: the mock server's rules and accepted
// proof schemes, plus the sequence of requests to make against it and the
// outcome each one is expected to produce.
type Scenario struct {
	// Name identifies the scenario in report output.
	Name string `json:"name"`
	// Rules are the protected routes the in-process mock server is
	// configured with for the duration of this scenario, in the same
	// shape as a serve --config file (see docs/PROTOCOL.md).
	Rules []mockserver.RuleConfig `json:"rules"`
	// HMACKey is the shared secret used to verify and sign hmac-sha256
	// proofs. Required only if a step uses that scheme.
	HMACKey string `json:"hmacKey,omitempty"`
	// Steps are the requests to make, in order, against the server.
	Steps []Step `json:"steps"`
}

// Step is one request/response exchange to drive through client.Loop and
// check against Expect.
type Step struct {
	// Name identifies the step in report output.
	Name string `json:"name"`
	// Method is the HTTP method to use. Empty defaults to GET.
	Method string `json:"method,omitempty"`
	// Path is the request path, relative to the scenario server's root.
	Path string `json:"path"`
	// Scheme selects the client.Signer used to settle a 402 challenge,
	// if one is issued. Empty defaults to the "fake" scheme. Unused if
	// the request is never challenged.
	Scheme string `json:"scheme,omitempty"`
	// Expect is the outcome this step's request must produce.
	Expect Expect `json:"expect"`
}

// Expect describes the required outcome of a Step, mirroring the fields
// of client.Result that matter for a pass/fail assertion.
type Expect struct {
	// Paid must match client.Result.Paid: whether a 402 challenge was
	// issued and a retry with a proof was attempted.
	Paid bool `json:"paid"`
	// FinalStatus must match client.Result.FinalStatusCode: the status
	// code of the last response received. Required (a zero value never
	// matches a real HTTP status and fails Validate).
	FinalStatus int `json:"finalStatus"`
}
