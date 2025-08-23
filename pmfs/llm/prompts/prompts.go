package prompts

import (
	"fmt"
	"strings"
)

// Prompt defines an interaction with a role-specific question and follow up.
type Prompt struct {
	ID       string
	Question string
	FollowUp string
}

// GetPrompts returns prompts for the given role or an error if the role is unknown.
func GetPrompts(role string) ([]Prompt, error) {
	switch strings.ToLower(role) {
	case "cto":
		return ctoPrompts, nil
	case "solution_architect":
		return solutionArchitectPrompts, nil
	case "qa_lead":
		return qaLeadPrompts, nil
	default:
		return nil, fmt.Errorf("unknown role: %s", role)
	}
}
