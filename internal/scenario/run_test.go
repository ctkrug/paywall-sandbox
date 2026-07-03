package scenario

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/ctkrug/paywall-sandbox/internal/client"
	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
)

func basicRules() []mockserver.RuleConfig {
	return []mockserver.RuleConfig{
		{Path: "/paid", Amount: 100, Asset: "USDC", Recipient: "0xsandbox"},
	}
}

func TestRunAllStepsPass(t *testing.T) {
	sc := Scenario{
		Name:  "happy path",
		Rules: basicRules(),
		Steps: []Step{
			{Name: "pays the fake scheme", Path: "/paid", Expect: Expect{Paid: true, FinalStatus: 200}},
			{Name: "unprotected route needs no payment", Path: "/free", Expect: Expect{Paid: false, FinalStatus: 200}},
		},
	}

	report, err := Run(context.Background(), sc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.Passed() {
		t.Fatalf("report.Passed() = false, want true; report:\n%s", report)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("len(report.Steps) = %d, want 2", len(report.Steps))
	}
}

func TestRunStepFailureIsRecordedNotFatal(t *testing.T) {
	sc := Scenario{
		Name:  "wrong expectation",
		Rules: basicRules(),
		Steps: []Step{
			{Name: "expects no payment but route is protected", Path: "/paid", Expect: Expect{Paid: false, FinalStatus: 200}},
			{Name: "correct expectation runs anyway", Path: "/free", Expect: Expect{Paid: false, FinalStatus: 200}},
		},
	}

	report, err := Run(context.Background(), sc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.Passed() {
		t.Fatal("report.Passed() = true, want false")
	}
	if report.Steps[0].Passed {
		t.Error("Steps[0].Passed = true, want false")
	}
	if !strings.Contains(report.Steps[0].Failure, "paid: want false, got true") {
		t.Errorf("Steps[0].Failure = %q, want it to mention the paid mismatch", report.Steps[0].Failure)
	}
	if !report.Steps[1].Passed {
		t.Error("Steps[1].Passed = false, want true; a failing step must not abort the scenario")
	}
}

func TestRunHMACSchemeRoundTrip(t *testing.T) {
	sc := Scenario{
		Name:    "hmac scheme",
		HMACKey: "shared-secret",
		Rules:   basicRules(),
		Steps: []Step{
			{Name: "pays with hmac-sha256", Path: "/paid", Scheme: "hmac-sha256", Expect: Expect{Paid: true, FinalStatus: 200}},
		},
	}

	report, err := Run(context.Background(), sc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.Passed() {
		t.Fatalf("report.Passed() = false, want true; report:\n%s", report)
	}
}

// TestRunHMACKeyMismatchRejected exercises a scenario capability plain
// pass/fail assertions can't from outside: a step where the client attempts
// payment (Paid=true) but the server rejects it (FinalStatus=402) because
// the signatures were computed with different keys. It builds the server
// directly (bypassing Scenario.Steps) since Scenario only carries one
// HMACKey shared by both sides.
func TestRunHMACKeyMismatchRejected(t *testing.T) {
	sc := Scenario{HMACKey: "server-secret", Rules: basicRules()}
	srv, err := sc.newServer()
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	defer srv.Close()

	loop := &client.Loop{Signer: client.HMACSigner{Key: []byte("wrong-secret")}}
	result, err := loop.Do(context.Background(), http.MethodGet, srv.URL+"/paid")
	if err != nil {
		t.Fatalf("loop.Do: %v", err)
	}
	if !result.Paid {
		t.Error("result.Paid = false, want true: a retry was attempted")
	}
	if result.FinalStatusCode != http.StatusPaymentRequired {
		t.Errorf("FinalStatusCode = %d, want 402: a mismatched signature must be rejected", result.FinalStatusCode)
	}
}

func TestRunUnknownScheme(t *testing.T) {
	sc := Scenario{
		Name:  "unknown scheme",
		Rules: basicRules(),
		Steps: []Step{
			{Name: "bogus scheme", Path: "/paid", Scheme: "bogus", Expect: Expect{Paid: true, FinalStatus: 200}},
		},
	}

	report, err := Run(context.Background(), sc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.Passed() {
		t.Fatal("report.Passed() = true, want false for an unknown scheme")
	}
	if report.Steps[0].Result != nil {
		t.Error("Steps[0].Result = non-nil, want nil since the request was never sent")
	}
}

func TestRunInvalidRulesFails(t *testing.T) {
	sc := Scenario{
		Name:  "bad rules",
		Rules: []mockserver.RuleConfig{{Path: "/paid"}}, // missing amount/asset/recipient
		Steps: []Step{
			{Name: "never runs", Path: "/paid", Expect: Expect{FinalStatus: 200}},
		},
	}

	if _, err := Run(context.Background(), sc); err == nil {
		t.Fatal("Run: want error for invalid rule set, got nil")
	}
}
