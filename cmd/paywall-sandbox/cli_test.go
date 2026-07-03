package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

func TestCLITestSubcommandMissingFileExitsNonZero(t *testing.T) {
	cmd := exec.Command(binPath, "test", "does-not-exist.json")
	if err := cmd.Run(); err == nil {
		t.Error("test does-not-exist.json: want non-zero exit, got success")
	}
}

func TestCLIRequestMissingURLExitsNonZero(t *testing.T) {
	cmd := exec.Command(binPath, "request")
	if err := cmd.Run(); err == nil {
		t.Error("request with no --url: want non-zero exit, got success")
	}
}
