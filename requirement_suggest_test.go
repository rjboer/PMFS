package PMFS

import (
	"fmt"
	"strings"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestRequirementSuggestOthers(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	r := Requirement{Description: "System shall X"}
	mockResp := `[{"name":"R2","description":"Desc2"},{"name":"R3","description":"Desc3"}]`
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		expected := fmt.Sprintf("Given the requirement %q", r.Description)
		if !strings.Contains(prompt, expected) {
			t.Fatalf("unexpected prompt: %s", prompt)
		}
		return mockResp, nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	prj := &ProjectType{ProductID: 1, ID: 1}
	reqs, err := r.SuggestOthers(prj)
	if err != nil {
		t.Fatalf("SuggestOthers: %v", err)
	}
	if len(reqs) != 2 || reqs[0].Name != "R2" || reqs[1].Description != "Desc3" {
		t.Fatalf("unexpected reqs: %#v", reqs)
	}
	if len(prj.D.PotentialRequirements) != 2 {
		t.Fatalf("requirements not appended: %#v", prj.D.PotentialRequirements)
	}
	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.PotentialRequirements) != 2 {
		t.Fatalf("requirements not persisted: %#v", prjReload.D.PotentialRequirements)
	}
}

func TestRequirementSuggestOthersMalformed(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	r := Requirement{Description: "System shall X"}
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return "not json", nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	prj := &ProjectType{ProductID: 1, ID: 1}
	if _, err := r.SuggestOthers(prj); err == nil {
		t.Fatalf("expected error for malformed response")
	}
}

func TestRequirementSuggestOthersCodeFence(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	r := Requirement{Description: "System shall X"}
	mockResp := "Sure!\n```json\n[{\"name\":\"R2\",\"description\":\"Desc2\"},{\"name\":\"R3\",\"description\":\"Desc3\"}]\n```"
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return mockResp, nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	prj := &ProjectType{ProductID: 1, ID: 1}
	reqs, err := r.SuggestOthers(prj)
	if err != nil {
		t.Fatalf("SuggestOthers: %v", err)
	}
	if len(reqs) != 2 || reqs[0].Name != "R2" || reqs[1].Description != "Desc3" {
		t.Fatalf("unexpected reqs: %#v", reqs)
	}
	if len(prj.D.PotentialRequirements) != 2 {
		t.Fatalf("requirements not appended: %#v", prj.D.PotentialRequirements)
	}
}
