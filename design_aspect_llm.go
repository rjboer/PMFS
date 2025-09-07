package PMFS

import (
	"encoding/json"
	"fmt"

	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

// GenerateTemplates asks the LLM for requirement templates related to this
// design aspect. Returned templates are appended to the aspect and also
// returned to the caller.
func (da *DesignAspect) GenerateTemplates(role, questionID string) ([]Requirement, error) {
	ps, err := prompts.GetPrompts(role)
	if err != nil {
		return nil, err
	}
	var p *prompts.Prompt
	for i := range ps {
		if ps[i].ID == questionID {
			p = &ps[i]
			break
		}
	}
	if p == nil {
		return nil, fmt.Errorf("prompt %s/%s not found", role, questionID)
	}
	prompt := fmt.Sprintf(p.Template, da.Description)
	resp, err := DB.LLM.Ask(prompt)
	if err != nil {
		return nil, err
	}
	raw, err := parseLLMJSON(resp)
	if err != nil {
		return nil, err
	}
	var reqs []Requirement
	if err := json.Unmarshal(raw, &reqs); err != nil {
		return nil, err
	}
	da.Templates = append(da.Templates, reqs...)
	return reqs, nil
}
