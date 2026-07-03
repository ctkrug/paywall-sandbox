package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ctkrug/paywall-sandbox/internal/client"
	"github.com/ctkrug/paywall-sandbox/internal/mockserver"
)

// Validate reports whether s is well-formed: it has at least one step,
// every step has a path and a required finalStatus, every hmac-sha256 step
// has an hmacKey to sign with, and Rules parses as a valid rule set.
func (s Scenario) Validate() error {
	if len(s.Steps) == 0 {
		return fmt.Errorf("scenario: at least one step is required")
	}
	if _, err := s.rules(); err != nil {
		return err
	}
	for i, st := range s.Steps {
		if st.Path == "" {
			return fmt.Errorf("scenario: step %d (%s): path must not be empty", i, st.Name)
		}
		if st.Expect.FinalStatus == 0 {
			return fmt.Errorf("scenario: step %d (%s): expect.finalStatus is required", i, st.Name)
		}
		if st.Scheme == client.HMACScheme && s.HMACKey == "" {
			return fmt.Errorf("scenario: step %d (%s): scheme %q requires a scenario-level hmacKey", i, st.Name, client.HMACScheme)
		}
	}
	return nil
}

// rules converts Rules to the mockserver.Rule slice Server expects,
// delegating to mockserver.LoadRules so a scenario's rule set is validated
// with exactly the same rules a serve --config file would be.
func (s Scenario) rules() ([]mockserver.Rule, error) {
	data, err := json.Marshal(mockserver.Config{Rules: s.Rules})
	if err != nil {
		return nil, fmt.Errorf("scenario: encoding rules: %w", err)
	}
	rules, err := mockserver.LoadRules(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("scenario: %w", err)
	}
	return rules, nil
}
