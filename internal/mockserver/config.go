package mockserver

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// RuleConfig is the JSON shape of one Rule in a rule set file. See
// docs/PROTOCOL.md and examples/rules.json for the format.
type RuleConfig struct {
	Method    string `json:"method,omitempty"`
	Path      string `json:"path"`
	Amount    uint64 `json:"amount"`
	Asset     string `json:"asset"`
	Recipient string `json:"recipient"`
}

// Config is the top-level shape of a rule set file: a list of protected
// routes and their payment terms.
type Config struct {
	Rules []RuleConfig `json:"rules"`
}

// LoadRules parses a Config from r and converts it to the Rules Server
// expects, validating every entry.
func LoadRules(r io.Reader) ([]Rule, error) {
	var cfg Config
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("mockserver: decoding rule config: %w", err)
	}

	rules := make([]Rule, 0, len(cfg.Rules))
	for i, rc := range cfg.Rules {
		rule := Rule{
			Method:    rc.Method,
			Path:      rc.Path,
			Amount:    rc.Amount,
			Asset:     rc.Asset,
			Recipient: rc.Recipient,
		}
		if err := validateRuleConfig(rule); err != nil {
			return nil, fmt.Errorf("mockserver: rule %d: %w", i, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// LoadRulesFile opens path and delegates to LoadRules.
func LoadRulesFile(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mockserver: opening rule config: %w", err)
	}
	defer f.Close()

	return LoadRules(f)
}

func validateRuleConfig(r Rule) error {
	if r.Path == "" {
		return fmt.Errorf("path must not be empty")
	}
	if r.Amount == 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if r.Asset == "" {
		return fmt.Errorf("asset must not be empty")
	}
	if r.Recipient == "" {
		return fmt.Errorf("recipient must not be empty")
	}
	return nil
}
