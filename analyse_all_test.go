package PMFS

import (
	"errors"
	"strings"
	"testing"

	llm "github.com/rjboer/PMFS/pmfs/llm"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

type mockClient struct{}

func (mockClient) AnalyzeAttachment(path string) ([]gemini.Requirement, error) { return nil, nil }

func (mockClient) Ask(prompt string) (string, error) {
	if strings.Contains(prompt, "fail") {
		return "", errors.New("ask error")
	}
	return "Yes", nil
}

func TestAnalyseAll(t *testing.T) {
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "q1", Template: "%s"}})
	defer prompts.SetTestPrompts(nil)

	orig := llm.SetClient(mockClient{})
	defer llm.SetClient(orig)

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

	prj.D.Requirements = []Requirement{{ID: 1, Description: "ok"}, {ID: 2, Description: "fail"}, {ID: 3, Description: "pot"}}

	err = prj.AnalyseAll("test", "q1", []string{"completeness-1"})
	if err == nil || err.Error() != "ask error" {
		t.Fatalf("expected ask error, got %v", err)
	}

	if len(prj.D.Requirements[0].GateResults) != 1 {
		t.Fatalf("expected gate results for requirement")
	}
	if len(prj.D.Requirements[2].GateResults) != 1 {
		t.Fatalf("expected gate results for third requirement")
	}

	var prjReload ProjectType
	prjReload.ID = prj.ID
	prjReload.ProductID = prj.ProductID
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Requirements[0].GateResults) != 1 {
		t.Fatalf("gate results not persisted")
	}
	if len(prjReload.D.Requirements[2].GateResults) != 1 {
		t.Fatalf("gate results for third requirement not persisted")
	}
}
