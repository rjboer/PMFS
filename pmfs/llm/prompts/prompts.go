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

// Prompt defines an interaction with a role-specific question and optional follow-up.
type Prompt struct {
	ID       string
	Question string
	FollowUp string // asked when the initial answer is "No"
}

// testPrompts holds prompts used for the special "test" role.
// It is populated by tests via SetTestPrompts and is ignored in normal use.
var testPrompts []Prompt

// SetTestPrompts registers prompts used when GetPrompts is called with role "test".
// It allows integration tests to supply deterministic questions and follow-ups.
func SetTestPrompts(ps []Prompt) { testPrompts = ps }

// GetPrompts returns prompts for the given role or an error if the role is unknown.
func GetPrompts(role string) ([]Prompt, error) {
	switch strings.ToLower(role) {
	case "cto":
		return ctoPrompts, nil
	case "devops_platform":
		return devOpsPlatformPrompts, nil
	case "ml_llm_engineer":
		return mlLlmEngineerPrompts, nil
	case "new_business_development":
		return newBusinessDevelopmentPrompts, nil
	case "product_manager":
		return productManagerPrompts, nil
	case "qa_lead":
		return qaLeadPrompts, nil
	case "sales":
		return salesPrompts, nil
	case "safety_compliance_lead":
		return safetyComplianceLeadPrompts, nil
	case "security_privacy_officer":
		return securityPrivacyOfficerPrompts, nil
	case "solution_architect":
		return solutionArchitectPrompts, nil
	case "ux_tech_writer":
		return uxTechWriterPrompts, nil
	case "test":
		if testPrompts == nil {
			return nil, fmt.Errorf("test prompts not set")
		}
		return testPrompts, nil
	default:
		return nil, fmt.Errorf("unknown role: %s", role)
	}
}
