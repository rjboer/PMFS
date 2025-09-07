package PMFS

import (
	"strings"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

func TestRequirementGenerateDesignAspects(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	r := Requirement{Description: "System shall X"}
	mockResp := `[{"name":"Aspect1","description":"Desc1"},{"name":"Aspect2","description":"Desc2"}]`
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		if !strings.Contains(prompt, r.Description) {
			t.Fatalf("unexpected prompt: %s", prompt)
		}
		return mockResp, nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	aspects, err := r.GenerateDesignAspects()
	if err != nil {
		t.Fatalf("GenerateDesignAspects: %v", err)
	}
	if len(aspects) != 2 || r.DesignAspects[0].Name != "Aspect1" || r.DesignAspects[1].Description != "Desc2" {
		t.Fatalf("unexpected aspects: %#v", r.DesignAspects)
	}
}

func TestProjectGenerateDesignAspectsAll(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.NewProduct(ProductData{Name: "prod"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	prd := &db.Products[0]
	if _, err := prd.NewProject(ProjectData{Name: "prj"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	prj := &prd.Projects[0]
	prj.D.Requirements = []Requirement{{ID: 1, Description: "ReqA"}, {ID: 2, Description: "ReqB"}}
	responses := map[string]string{
		"ReqA": `[{"name":"A1","description":"D1"}]`,
		"ReqB": `[{"name":"B1","description":"D2"}]`,
	}
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		for k, v := range responses {
			if strings.Contains(prompt, k) {
				return v, nil
			}
		}
		t.Fatalf("unexpected prompt: %s", prompt)
		return "", nil
	}}
	DB.LLM = client
	if err := prj.GenerateDesignAspectsAll(); err != nil {
		t.Fatalf("GenerateDesignAspectsAll: %v", err)
	}
	if len(prj.D.Requirements[0].DesignAspects) != 1 || prj.D.Requirements[0].DesignAspects[0].Name != "A1" {
		t.Fatalf("aspects not stored: %#v", prj.D.Requirements[0].DesignAspects)
	}
	var prjReload ProjectType
	prjReload.ID = prj.ID
	prjReload.ProductID = prj.ProductID
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Requirements[1].DesignAspects) != 1 || prjReload.D.Requirements[1].DesignAspects[0].Name != "B1" {
		t.Fatalf("aspects not persisted: %#v", prjReload.D.Requirements[1].DesignAspects)
	}
}

func TestDesignAspectGenerateTemplates(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "1", Template: "Given the design aspect %s, list requirement templates"}})
	da := DesignAspect{Description: "Improve security"}
	mockResp := `[{"name":"Template1","description":"System shall enforce XX"}]`
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		if !strings.Contains(prompt, da.Description) {
			t.Fatalf("unexpected prompt: %s", prompt)
		}
		return mockResp, nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	reqs, err := da.GenerateTemplates("test", "1")
	if err != nil {
		t.Fatalf("GenerateTemplates: %v", err)
	}
	if len(reqs) != 1 || da.Templates[0].Name != "Template1" {
		t.Fatalf("unexpected templates: %#v", da.Templates)
	}
}

func TestDesignAspectGenerateTemplatesMalformed(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "1", Template: "Given the design aspect %s"}})
	da := DesignAspect{Description: "Improve reliability"}
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) { return "not json", nil }}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	if _, err := da.GenerateTemplates("test", "1"); err == nil {
		t.Fatalf("expected error for malformed response")
	}
}

func TestDesignAspectEvaluateDesignGates(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	da := DesignAspect{Templates: []Requirement{{Description: "Req1"}, {Description: "Req2"}}}
	client := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		return "Yes", nil
	}}
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	DB.LLM = client
	if err := da.EvaluateDesignGates([]string{"clarity-form-1"}); err != nil {
		t.Fatalf("EvaluateDesignGates: %v", err)
	}
	for i := range da.Templates {
		if len(da.Templates[i].GateResults) != 1 || !da.Templates[i].GateResults[0].Pass {
			t.Fatalf("gate results not stored: %#v", da.Templates[i].GateResults)
		}
	}
}
