package PMFS

import (
	"fmt"
	"strings"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestRequirementSuggestOthers(t *testing.T) {
	r := Requirement{Description: "System shall X"}
	mockResp := `[{"name":"R2","description":"Desc2"},{"name":"R3","description":"Desc3"}]`
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		expected := fmt.Sprintf("Given the requirement %q", r.Description)
		if !strings.Contains(prompt, expected) {
			t.Fatalf("unexpected prompt: %s", prompt)
		}
		return mockResp, nil
	}}
	db := &Database{LLM: client}
	reqs, err := r.SuggestOthers(db)
	if err != nil {
		t.Fatalf("SuggestOthers: %v", err)
	}
	if len(reqs) != 2 || reqs[0].Name != "R2" || reqs[1].Description != "Desc3" {
		t.Fatalf("unexpected reqs: %#v", reqs)
	}
}

func TestRequirementSuggestOthersMalformed(t *testing.T) {
	r := Requirement{Description: "System shall X"}
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return "not json", nil
	}}
	db := &Database{LLM: client}
	if _, err := r.SuggestOthers(db); err == nil {
		t.Fatalf("expected error for malformed response")
	}
}
