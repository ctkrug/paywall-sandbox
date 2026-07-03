package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
)

// binPath is set by TestMain to a binary built once for the whole package,
// so black-box subcommand tests don't each pay a `go build`.
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "paywall-sandbox-cli-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	binPath = filepath.Join(dir, "paywall-sandbox")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic("build cli test binary: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

func TestCLINoArgsExitsNonZero(t *testing.T) {
	cmd := exec.Command(binPath)
	if err := cmd.Run(); err == nil {
		t.Error("running with no args: want non-zero exit, got success")
	}
}

func TestCLIVersionExitsZero(t *testing.T) {
	cmd := exec.Command(binPath, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version: %v (%s)", err, out)
	}
	if !strings.Contains(string(out), "paywall-sandbox") {
		t.Errorf("version output = %q, want it to contain %q", out, "paywall-sandbox")
	}
}

func TestCLITestSubcommandRunsExampleScenario(t *testing.T) {
	cmd := exec.Command(binPath, "test", "../../examples/scenario.json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test examples/scenario.json: %v (%s)", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Errorf("test output = %q, want it to contain PASS", out)
	}
}

func TestCLITestSubcommandVerboseTracesSteps(t *testing.T) {
	cmd := exec.Command(binPath, "test", "--verbose", "../../examples/scenario.json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test --verbose examples/scenario.json: %v (%s)", err, out)
	}
	if !strings.Contains(string(out), "descriptor:") {
		t.Errorf("test --verbose output = %q, want it to include a descriptor trace", out)
	}
}

func TestCLITestSubcommandRunsHMACScenario(t *testing.T) {
	cmd := exec.Command(binPath, "test", "../../examples/scenario-hmac.json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test examples/scenario-hmac.json: %v (%s)", err, out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Errorf("test output = %q, want it to contain PASS", out)
	}
}

func TestCLITestSubcommandMissingFileExitsNonZero(t *testing.T) {
	cmd := exec.Command(binPath, "test", "does-not-exist.json")
	if err := cmd.Run(); err == nil {
		t.Error("test does-not-exist.json: want non-zero exit, got success")
	}
}

func TestCLITestSubcommandFailingStepExitsNonZero(t *testing.T) {
	const failingScenario = `{
		"name": "deliberately wrong expectation",
		"rules": [{"path": "/paid", "amount": 100, "asset": "USDC", "recipient": "0xsandbox"}],
		"steps": [
			{"name": "expects a status the server will never return", "method": "GET", "path": "/paid", "expect": {"paid": true, "finalStatus": 599}}
		]
	}`
	path := filepath.Join(t.TempDir(), "failing.json")
	if err := os.WriteFile(path, []byte(failingScenario), 0o644); err != nil {
		t.Fatalf("write scenario: %v", err)
	}

	cmd := exec.Command(binPath, "test", path)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("test %s: want non-zero exit, got success (%s)", path, out)
	}
	if !strings.Contains(string(out), "FAIL") {
		t.Errorf("test output = %q, want it to contain FAIL", out)
	}
}

func TestCLIRequestMissingURLExitsNonZero(t *testing.T) {
	cmd := exec.Command(binPath, "request")
	if err := cmd.Run(); err == nil {
		t.Error("request with no --url: want non-zero exit, got success")
	}
}

func TestCLIRequestSettlesAgainstMockServer(t *testing.T) {
	srv := httptest.NewServer(&mockserver.Server{
		Rules: []mockserver.Rule{{Path: "/paid", Amount: 100, Asset: "USDC", Recipient: "0xsandbox"}},
		Next: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	})
	defer srv.Close()

	cmd := exec.Command(binPath, "request", "--url", srv.URL+"/paid", "--verbose")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("request --url %s: %v (%s)", srv.URL+"/paid", err, out)
	}
	if !strings.Contains(string(out), "-> 200") {
		t.Errorf("request output = %q, want it to contain %q", out, "-> 200")
	}
	if !strings.Contains(string(out), "proof:") {
		t.Errorf("request --verbose output = %q, want it to include the proof trace", out)
	}
}
