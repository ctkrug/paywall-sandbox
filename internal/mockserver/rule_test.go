package mockserver

import (
	"net/http/httptest"
	"testing"
)

func TestRuleMatches(t *testing.T) {
	cases := []struct {
		name   string
		rule   Rule
		method string
		path   string
		want   bool
	}{
		{"exact match", Rule{Method: "GET", Path: "/paid"}, "GET", "/paid", true},
		{"method case-insensitive", Rule{Method: "get", Path: "/paid"}, "GET", "/paid", true},
		{"wildcard method", Rule{Path: "/paid"}, "POST", "/paid", true},
		{"wrong method", Rule{Method: "GET", Path: "/paid"}, "POST", "/paid", false},
		{"wrong path", Rule{Method: "GET", Path: "/paid"}, "GET", "/free", false},
		{"wildcard matches prefix itself", Rule{Path: "/api/*"}, "GET", "/api", true},
		{"wildcard matches nested path", Rule{Path: "/api/*"}, "GET", "/api/v1/widgets", true},
		{"wildcard rejects sibling prefix", Rule{Path: "/api/*"}, "GET", "/apikeys", false},
		{"wildcard rejects unrelated path", Rule{Path: "/api/*"}, "GET", "/free", false},
		{"bare wildcard protects every path", Rule{Path: "/*"}, "GET", "/anything/nested", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if got := tc.rule.Matches(req); got != tc.want {
				t.Fatalf("Matches() = %v, want %v", got, tc.want)
			}
		})
	}
}
