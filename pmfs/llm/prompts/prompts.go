// Package prompts defines role-specific questions for the PMFS library.
//
// Each Prompt may include a FollowUp string that is sent verbatim when the
// initial answer is "No". Follow-up text should be a standalone question so
// that it can be provided directly to the LLM's Ask function.
package prompts

import (
	"fmt"
	"strings"
)

// Prompt defines an interaction with a role-specific question template and optional follow-up.
type Prompt struct {
	ID       string
	Template string
	FollowUp string // asked when the initial answer is "No"
}

// testPrompts holds prompts used for the special "test" role.
// It is populated by tests via SetTestPrompts and is ignored in normal use.
var testPrompts []Prompt

// rolePrompts maps a role name to its registered prompts.
var rolePrompts = map[string][]Prompt{}

// RegisterRole registers prompts for a given role. Role names are stored in
// lowercase to ensure case-insensitive lookups.
func RegisterRole(role string, prompts []Prompt) {
	rolePrompts[strings.ToLower(role)] = prompts
}

// SetTestPrompts registers prompts used when GetPrompts is called with role "test".
// It allows integration tests to supply deterministic questions and follow-ups.
func SetTestPrompts(ps []Prompt) { testPrompts = ps }

// GetPrompts returns prompts for the given role or an error if the role is unknown.
func GetPrompts(role string) ([]Prompt, error) {
	r := strings.ToLower(role)
	if r == "test" {
		if testPrompts == nil {
			return nil, fmt.Errorf("test prompts not set")
		}
		return testPrompts, nil
	}
	ps, ok := rolePrompts[r]
	if !ok {
		return nil, fmt.Errorf("unknown role: %s", role)
	}
	return ps, nil
}
