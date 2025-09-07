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
	reqs := []Requirement{{Name: "R1", Description: "same"}, {Name: "R2", Description: "same"}}
	out := Deduplicate(reqs)
	if len(out) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(out))
	}
}
