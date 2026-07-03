package mockserver

import (
	"strings"
	"testing"
)

func TestLoadRulesValidConfig(t *testing.T) {
	const src = `{
		"rules": [
			{"path": "/paid", "amount": 100, "asset": "USDC", "recipient": "0xsandbox"},
			{"method": "POST", "path": "/api/*", "amount": 250, "asset": "USD", "recipient": "acct-1"}
		]
	}`

	rules, err := LoadRules(strings.NewReader(src))
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(rules))
	}
	if rules[0].Path != "/paid" || rules[0].Amount != 100 || rules[0].Asset != "USDC" {
		t.Fatalf("rules[0] = %+v", rules[0])
	}
	if rules[1].Method != "POST" || rules[1].Path != "/api/*" {
		t.Fatalf("rules[1] = %+v", rules[1])
	}
}

func TestLoadRulesInvalidJSON(t *testing.T) {
	if _, err := LoadRules(strings.NewReader("not json")); err == nil {
		t.Fatal("LoadRules() error = nil, want decode error")
	}
}

func TestLoadRulesRejectsMissingFields(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"empty path", `{"rules":[{"path":"","amount":100,"asset":"USD","recipient":"r"}]}`},
		{"zero amount", `{"rules":[{"path":"/p","amount":0,"asset":"USD","recipient":"r"}]}`},
		{"empty asset", `{"rules":[{"path":"/p","amount":100,"asset":"","recipient":"r"}]}`},
		{"empty recipient", `{"rules":[{"path":"/p","amount":100,"asset":"USD","recipient":""}]}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := LoadRules(strings.NewReader(tc.src)); err == nil {
				t.Fatal("LoadRules() error = nil, want validation error")
			}
		})
	}
}

func TestLoadRulesFileMissing(t *testing.T) {
	if _, err := LoadRulesFile("/nonexistent/rules.json"); err == nil {
		t.Fatal("LoadRulesFile() error = nil, want open error")
	}
}
