package PMFS

import (
	"encoding/json"
	"fmt"
)

// SuggestOthers asks the client for related potential requirements based on
// this requirement's description. Returned requirements are appended to the
// project (if provided) and persisted immediately.
func (r *Requirement) SuggestOthers(prj *ProjectType) ([]Requirement, error) {
	prompt := fmt.Sprintf("Given the requirement %q, list other potential requirements (JSON array with `name` and `description`).", r.Description)
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
	reqs = Deduplicate(reqs)
	if prj != nil {
		prj.D.PotentialRequirements = Deduplicate(append(prj.D.PotentialRequirements, reqs...))
		if err := prj.Save(); err != nil {
			return nil, err
		}
	}
	return reqs, nil
}

// GenerateDesignAspects asks the client for design improvement topics based on
// the requirement's description. Returned aspects are appended to the
// requirement and also returned to the caller.
func (r *Requirement) GenerateDesignAspects() ([]DesignAspect, error) {
	prompt := fmt.Sprintf("Given the requirement %q, list design improvement topics (JSON array with `name` and `description`).", r.Description)
	resp, err := DB.LLM.Ask(prompt)
	if err != nil {
		return nil, err
	}
	raw, err := parseLLMJSON(resp)
	if err != nil {
		return nil, err
	}
	var aspects []DesignAspect
	if err := json.Unmarshal(raw, &aspects); err != nil {
		return nil, err
	}
	r.DesignAspects = append(r.DesignAspects, aspects...)
	return aspects, nil
}
