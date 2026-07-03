package scenario

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Load parses a Scenario from r and validates it.
func Load(r io.Reader) (Scenario, error) {
	var sc Scenario
	if err := json.NewDecoder(r).Decode(&sc); err != nil {
		return Scenario{}, fmt.Errorf("scenario: decoding: %w", err)
	}
	if err := sc.Validate(); err != nil {
		return Scenario{}, err
	}
	return sc, nil
}

// LoadFile opens path and delegates to Load.
func LoadFile(path string) (Scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("scenario: opening %s: %w", path, err)
	}
	defer f.Close()

	return Load(f)
}
