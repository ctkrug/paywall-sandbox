// Package mockserver implements a local HTTP server that challenges
// configured routes with a 402 Payment Required response and serves them
// once a matching proof of payment is presented.
package mockserver

import (
	"net/http"
	"strings"
)

// Rule describes one route that requires payment before it is served, and
// the terms of that payment.
type Rule struct {
	// Method restricts the rule to a single HTTP method. Empty matches any
	// method.
	Method string
	// Path is the exact request path this rule applies to.
	Path string
	// Amount is the price in the smallest unit of Asset.
	Amount uint64
	// Asset identifies the currency or token, e.g. "USD" or "USDC".
	Asset string
	// Recipient is the address or account payment must settle to.
	Recipient string
}

// Matches reports whether r applies to req.
func (r Rule) Matches(req *http.Request) bool {
	if r.Method != "" && !strings.EqualFold(r.Method, req.Method) {
		return false
	}
	return r.Path == req.URL.Path
}
