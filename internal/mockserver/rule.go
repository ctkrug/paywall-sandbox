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
	// Path is the request path this rule applies to. A trailing "/*"
	// matches the prefix itself or anything nested under it (e.g. "/api/*"
	// matches "/api", "/api/foo", and "/api/foo/bar"); anything else must
	// match the request path exactly.
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
	return matchPath(r.Path, req.URL.Path)
}

// matchPath reports whether path satisfies pattern. A pattern ending in
// "/*" matches the prefix itself or anything nested under it; any other
// pattern must match path exactly.
func matchPath(pattern, path string) bool {
	prefix, ok := strings.CutSuffix(pattern, "/*")
	if !ok {
		return pattern == path
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
