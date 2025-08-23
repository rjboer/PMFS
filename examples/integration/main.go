package main

import (
	"fmt"
	"log"
	"strings"

	PMFS "github.com/rjboer/PMFS"
	"github.com/rjboer/PMFS/pmfs/llm/gates"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/interact"
)

// This example demonstrates a full flow using Gemini to analyze a document,
// storing the returned requirement, asking a role-specific question about it,
// and finally evaluating it against quality gates.
func main() {
	// Stub the Gemini client so the example runs without external calls.
	stub := gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return []gemini.Requirement{{ID: 1, Name: "Login", Description: "Users shall log in with email and password."}}, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if strings.Contains(strings.ToLower(prompt), "answer yes or no only") {
				return "Yes", nil
			}
			if strings.Contains(strings.ToLower(prompt), "given the requirement") {
				return "Yes", nil
			}
			return "stub response", nil
		},
	}
	prev := gemini.SetClient(stub)
	defer gemini.SetClient(prev)

	// Analyze a document to extract potential requirements.
	reqs, err := gemini.AnalyzeAttachment("testdata/spec1.txt")
	if err != nil {
		log.Fatalf("analyze: %v", err)
	}
	if len(reqs) == 0 {
		log.Fatal("no requirements returned")
	}

	// Store the first requirement in a project structure.
	prj := PMFS.ProjectType{}
	prj.D.PotentialRequirements = append(prj.D.PotentialRequirements, PMFS.Requirement{
		Name:        reqs[0].Name,
		Description: reqs[0].Description,
	})
	r := prj.D.PotentialRequirements[0]
	fmt.Printf("Requirement: %s - %s\n", r.Name, r.Description)

	// Ask a QA lead if the requirement's testing strategy is acceptable.
	pass, follow, err := interact.RunQuestion(stub, "qa_lead", "1", r.Description)
	if err != nil {
		log.Fatalf("run question: %v", err)
	}
	fmt.Printf("QA Lead agrees? %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}

	// Evaluate the requirement against quality gates.
	gateResults, err := gates.Evaluate(stub, []string{"clarity-form-1", "duplicate-1"}, r.Description)
	if err != nil {
		log.Fatalf("gates evaluate: %v", err)
	}
	for _, gr := range gateResults {
		fmt.Printf("Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
	}
}
