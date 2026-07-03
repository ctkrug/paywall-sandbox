package scenario

import (
	"strings"
	"testing"
)

func TestReportPassed(t *testing.T) {
	passing := Report{Steps: []StepResult{{Passed: true}, {Passed: true}}}
	if !passing.Passed() {
		t.Error("Passed() = false, want true when every step passed")
	}

	mixed := Report{Steps: []StepResult{{Passed: true}, {Passed: false}}}
	if mixed.Passed() {
		t.Error("Passed() = true, want false when a step failed")
	}
}

func TestReportString(t *testing.T) {
	r := Report{
		ScenarioName: "checkout",
		Steps: []StepResult{
			{Name: "pays", Passed: true},
			{Name: "rejects stale nonce", Passed: false, Failure: "finalStatus: want 402, got 200"},
		},
	}

	got := r.String()
	for _, want := range []string{"checkout", "[PASS] pays", "[FAIL] rejects stale nonce", "finalStatus: want 402, got 200"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}
