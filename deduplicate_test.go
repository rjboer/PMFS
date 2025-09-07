package PMFS

import (
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestDeduplicate(t *testing.T) {
	stub := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return "1", nil
	}}
	orig := DB
	DB = &Database{LLM: stub}
	defer func() { DB = orig }()

	reqs := []Requirement{
		{Name: "R1", Description: "System shall log in"},
		{Name: "R1 copy", Description: "System shall log in"},
	}
	deduped := Deduplicate(reqs)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(deduped))
	}
}
