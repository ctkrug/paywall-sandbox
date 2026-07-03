package main

import "testing"

func TestLoadServeRulesDefaultsToFlags(t *testing.T) {
	rules, err := loadServeRules("", "/paid", 100, "USDC", "0xsandbox")
	if err != nil {
		t.Fatalf("loadServeRules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(rules))
	}
	rule := rules[0]
	if rule.Path != "/paid" || rule.Amount != 100 || rule.Asset != "USDC" || rule.Recipient != "0xsandbox" {
		t.Errorf("rules[0] = %+v, want {/paid 100 USDC 0xsandbox ...}", rule)
	}
}

func TestLoadServeRulesFromConfig(t *testing.T) {
	rules, err := loadServeRules("../../examples/rules.json", "/paid", 100, "USDC", "0xsandbox")
	if err != nil {
		t.Fatalf("loadServeRules: %v", err)
	}
	if len(rules) < 1 {
		t.Fatalf("len(rules) = %d, want at least 1", len(rules))
	}
}

func TestLoadServeRulesConfigNotFound(t *testing.T) {
	if _, err := loadServeRules("does-not-exist.json", "/paid", 100, "USDC", "0xsandbox"); err == nil {
		t.Fatal("loadServeRules: want error for missing config file, got nil")
	}
}
