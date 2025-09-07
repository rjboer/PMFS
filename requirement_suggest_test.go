package PMFS

import (
	"path/filepath"
	"strings"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestRequirementSuggestOthers(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	prj := &ProjectType{ProductID: 1, ID: 1}
	prj.D.Requirements = []Requirement{{Description: "System shall X"}}
	r := &prj.D.Requirements[0]
	mockResp := `[{"name":"R2","description":"Desc2"},{"name":"Dup","description":"System shall X"}]`
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		if strings.Contains(prompt, "Given the requirement") {
			return mockResp, nil
		}
		if strings.Count(prompt, "System shall X") >= 2 {
			return "yes", nil
		}
		return "no", nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	reqs, err := r.SuggestOthers(prj)
	if err != nil {
		t.Fatalf("SuggestOthers: %v", err)
	}
	if len(reqs) != 2 || reqs[0].Name != "R2" || reqs[1].Name != "Dup" {
		t.Fatalf("unexpected reqs: %#v", reqs)
	}
	if len(prj.D.Requirements) != 2 {
		t.Fatalf("requirements not deduplicated: %#v", prj.D.Requirements)
	}
	if prj.D.Requirements[1].ParentID != 0 {
		t.Fatalf("parent index not set: %#v", prj.D.Requirements[1])
	}
	if !prj.D.Requirements[1].Condition.Proposed || prj.D.Requirements[1].Condition.AIgenerated {
		t.Fatalf("condition flags not set correctly: %#v", prj.D.Requirements[1].Condition)
	}
	var dp struct {
		D ProjectData `toml:"projectdata"`
	}
	path := filepath.Join(projectDir(prj.ProductID, prj.ID), projectTOML)
	if err := readTOML(path, &dp); err != nil {
		t.Fatalf("readTOML: %v", err)
	}
	if len(dp.D.Requirements) != 2 {
		t.Fatalf("project.toml not updated: %#v", dp.D.Requirements)
	}
}

func TestRequirementSuggestOthersMalformed(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	prj := &ProjectType{ProductID: 1, ID: 1}
	prj.D.Requirements = []Requirement{{Description: "System shall X"}}
	r := &prj.D.Requirements[0]
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return "not json", nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	if _, err := r.SuggestOthers(prj); err == nil {
		t.Fatalf("expected error for malformed response")
	}
}

func TestRequirementSuggestOthersCodeFence(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	prj := &ProjectType{ProductID: 1, ID: 1}
	prj.D.Requirements = []Requirement{{Description: "System shall X"}}
	r := &prj.D.Requirements[0]
	mockResp := "Sure!\n```json\n[{\"name\":\"R2\",\"description\":\"Desc2\"},{\"name\":\"R3\",\"description\":\"Desc3\"}]\n```"
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		if strings.Contains(prompt, "Given the requirement") {
			return mockResp, nil
		}
		return "no", nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	reqs, err := r.SuggestOthers(prj)
	if err != nil {
		t.Fatalf("SuggestOthers: %v", err)
	}
	if len(reqs) != 2 || reqs[0].Name != "R2" || reqs[1].Description != "Desc3" {
		t.Fatalf("unexpected reqs: %#v", reqs)
	}
	if len(prj.D.Requirements) != 3 {
		t.Fatalf("requirements not appended: %#v", prj.D.Requirements)
	}
	if prj.D.Requirements[1].ParentID != 0 {
		t.Fatalf("parent index not set: %#v", prj.D.Requirements[1])
	}

	var dp2 struct {
		D ProjectData `toml:"projectdata"`
	}
	path := filepath.Join(projectDir(prj.ProductID, prj.ID), projectTOML)
	if err := readTOML(path, &dp2); err != nil {
		t.Fatalf("readTOML: %v", err)
	}
	if len(dp2.D.Requirements) != 3 {
		t.Fatalf("project.toml not updated: %#v", dp2.D.Requirements)
	}

}
