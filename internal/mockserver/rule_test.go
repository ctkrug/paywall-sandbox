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
