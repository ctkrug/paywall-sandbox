package scenario

import (
	"testing"

	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
)

func TestValidateNoSteps(t *testing.T) {
	sc := Scenario{Name: "empty"}
	if err := sc.Validate(); err == nil {
		t.Fatal("Validate: want error for scenario with no steps, got nil")
	}
}

func TestValidateStepMissingPath(t *testing.T) {
	sc := Scenario{
		Steps: []Step{{Name: "missing path", Expect: Expect{FinalStatus: 200}}},
	}
	if err := sc.Validate(); err == nil {
		t.Fatal("Validate: want error for step with no path, got nil")
	}
}

func TestValidateStepMissingFinalStatus(t *testing.T) {
	sc := Scenario{
		Steps: []Step{{Name: "no status", Path: "/paid"}},
	}
	if err := sc.Validate(); err == nil {
		t.Fatal("Validate: want error for step missing expect.finalStatus, got nil")
	}
}

func TestValidateHMACStepWithoutKey(t *testing.T) {
	sc := Scenario{
		Steps: []Step{{
			Name:   "hmac without key",
			Path:   "/paid",
			Scheme: "hmac-sha256",
			Expect: Expect{FinalStatus: 200},
		}},
	}
	if err := sc.Validate(); err == nil {
		t.Fatal("Validate: want error for hmac-sha256 step with no hmacKey, got nil")
	}
}

func TestValidateHMACStepWithKey(t *testing.T) {
	sc := Scenario{
		HMACKey: "shared-secret",
		Steps: []Step{{
			Name:   "hmac with key",
			Path:   "/paid",
			Scheme: "hmac-sha256",
			Expect: Expect{Paid: true, FinalStatus: 200},
		}},
	}
	if err := sc.Validate(); err != nil {
		t.Errorf("Validate: unexpected error: %v", err)
	}
}

func TestValidateBadRule(t *testing.T) {
	sc := Scenario{
		Rules: []mockserver.RuleConfig{{Path: "/paid"}}, // amount/asset/recipient missing
		Steps: []Step{{Name: "bad rule", Path: "/paid", Expect: Expect{FinalStatus: 200}}},
	}
	if err := sc.Validate(); err == nil {
		t.Fatal("Validate: want error for invalid rule config, got nil")
	}
}
