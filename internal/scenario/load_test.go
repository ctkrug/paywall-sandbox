package scenario

import (
	"strings"
	"testing"
)

func validScenarioJSON() string {
	return `{
		"name": "basic",
		"rules": [{"path": "/paid", "amount": 100, "asset": "USDC", "recipient": "0xsandbox"}],
		"steps": [
			{"name": "pays", "path": "/paid", "expect": {"paid": true, "finalStatus": 200}}
		]
	}`
}

func TestLoadValid(t *testing.T) {
	sc, err := Load(strings.NewReader(validScenarioJSON()))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if sc.Name != "basic" {
		t.Errorf("Name = %q, want %q", sc.Name, "basic")
	}
	if len(sc.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(sc.Steps))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	if _, err := Load(strings.NewReader("{not json")); err == nil {
		t.Fatal("Load: want error for malformed JSON, got nil")
	}
}

func TestLoadInvalidScenario(t *testing.T) {
	if _, err := Load(strings.NewReader(`{"name": "empty"}`)); err == nil {
		t.Fatal("Load: want error for scenario with no steps, got nil")
	}
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile("testdata/does-not-exist.json"); err == nil {
		t.Fatal("LoadFile: want error for missing file, got nil")
	}
}

func TestLoadFile(t *testing.T) {
	sc, err := LoadFile("testdata/basic.json")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if sc.Name != "basic" {
		t.Errorf("Name = %q, want %q", sc.Name, "basic")
	}
}
