package PMFS

import (
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestDeduplicate(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return "yes", nil
	}}
	reqs := []Requirement{
		{Name: "R1", Description: "same"},
		{Name: "R2", Description: "same"},
		{Name: "R3", Description: "same", Condition: ConditionType{Proposed: true}},
		{Name: "R4", Description: "other", Condition: ConditionType{Deleted: true}},
	}

	out := Deduplicate(reqs, true)
	if len(out) != 2 {
		t.Fatalf("expected 2 requirements when ignoring proposed, got %d", len(out))
	}
	if !out[1].Condition.Proposed {
		t.Fatalf("expected proposed requirement to remain")
	}

	out = Deduplicate(reqs, false)
	if len(out) != 1 {
		t.Fatalf("expected 1 requirement when including proposed, got %d", len(out))
	}
}
