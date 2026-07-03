package scenario

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ctkrug/paywall-sandbox/internal/client"
)

// Run executes every Step in s in order against a freshly started
// in-process mock server, and returns a Report of what happened. A non-nil
// error means the scenario itself couldn't be set up (e.g. an invalid rule
// set); a step whose request or assertions fail is instead recorded as a
// failing StepResult, so Run always finishes the scenario.
func Run(ctx context.Context, s Scenario) (Report, error) {
	srv, err := s.newServer()
	if err != nil {
		return Report{}, err
	}
	defer srv.Close()

	report := Report{ScenarioName: s.Name}
	for _, step := range s.Steps {
		report.Steps = append(report.Steps, s.runStep(ctx, srv.URL, step))
	}
	return report, nil
}

func (s Scenario) runStep(ctx context.Context, baseURL string, step Step) StepResult {
	method := step.Method
	if method == "" {
		method = http.MethodGet
	}

	signer, err := signerFor(step.Scheme, []byte(s.HMACKey))
	if err != nil {
		return StepResult{Name: step.Name, Failure: err.Error()}
	}

	loop := &client.Loop{Signer: signer}
	result, err := loop.Do(ctx, method, baseURL+step.Path)
	if err != nil {
		return StepResult{Name: step.Name, Failure: fmt.Sprintf("request failed: %v", err)}
	}

	if failure := checkExpect(step.Expect, result); failure != "" {
		return StepResult{Name: step.Name, Failure: failure, Result: result}
	}
	return StepResult{Name: step.Name, Passed: true, Result: result}
}

// checkExpect compares result against exp, returning a description of
// every mismatch, or "" if exp is fully satisfied.
func checkExpect(exp Expect, result *client.Result) string {
	var problems []string
	if result.Paid != exp.Paid {
		problems = append(problems, fmt.Sprintf("paid: want %v, got %v", exp.Paid, result.Paid))
	}
	if result.FinalStatusCode != exp.FinalStatus {
		problems = append(problems, fmt.Sprintf("finalStatus: want %d, got %d", exp.FinalStatus, result.FinalStatusCode))
	}
	return strings.Join(problems, "; ")
}
