package scenario

import (
	"fmt"
	"strings"

	"github.com/ctkrug/paywall-sandbox/internal/client"
)

// StepResult is the outcome of running one Step.
type StepResult struct {
	Name string
	// Passed reports whether the step's actual outcome matched its
	// Expect. Failure is empty when Passed is true.
	Passed  bool
	Failure string
	// Result is the client.Loop trace for this step, nil only if the
	// request itself could not be sent (e.g. an unknown scheme).
	Result *client.Result
}

// Report is the outcome of running every Step in a Scenario, in order.
type Report struct {
	ScenarioName string
	Steps        []StepResult
}

// Passed reports whether every step in the report passed.
func (r Report) Passed() bool {
	for _, s := range r.Steps {
		if !s.Passed {
			return false
		}
	}
	return true
}

// String renders a human-readable pass/fail summary, suitable for direct
// CLI output.
func (r Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "scenario: %s\n", r.ScenarioName)
	for _, s := range r.Steps {
		status := "PASS"
		if !s.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(&b, "  [%s] %s\n", status, s.Name)
		if !s.Passed {
			fmt.Fprintf(&b, "        %s\n", s.Failure)
		}
	}
	return b.String()
}
