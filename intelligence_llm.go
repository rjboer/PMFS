package PMFS

import (
	"encoding/json"
	"fmt"
)

// summarizeContent asks the LLM to summarize the given content.
func summarizeContent(content string) (string, error) {
	prompt := fmt.Sprintf("Summarize the following content:\n%s", content)
	return DB.LLM.Ask(prompt)
}

// designAspectsFromSummary asks the LLM for design improvement topics based on the summary.
func designAspectsFromSummary(summary string) ([]DesignAspect, error) {
	prompt := fmt.Sprintf("Given the intelligence summary %q, list design improvement topics (JSON array with `name` and `description`).", summary)
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
	return aspects, nil
}
