package main

import (
	"fmt"
	"log"
	"strings"

	PMFS "github.com/rjboer/PMFS"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// This example demonstrates a full flow using Gemini to analyze a document,
// storing the returned requirement, asking a role-specific question about it,
// and finally evaluating it against quality gates. Once the Gemini client is
// configured, requirement methods like Analyse and EvaluateGates can be called
// directly without additional setup. Remove the stub below and set
// GEMINI_API_KEY to exercise the real API.
func main() {
	// Stub the Gemini client so the example runs without external calls.
	// Remove this block and set GEMINI_API_KEY to use the real API via
	// the default client.
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
	PMFS.SetBaseDir(".")
	prj := PMFS.ProjectType{ProductID: 0, ID: 0}
	att := PMFS.Attachment{RelPath: "../../../testdata/spec1.txt"}

	// Analyze a document to extract potential requirements.
	if err := att.Analyze(&prj); err != nil {
		log.Fatalf("analyze: %v", err)
	}
	if len(prj.D.PotentialRequirements) == 0 {
		log.Fatal("no requirements returned")
	}

	r := &prj.D.PotentialRequirements[0]

	fmt.Printf("Requirement: %s - %s\n", r.Name, r.Description)

	// With the client configured above, the requirement can query roles and
	// evaluate gates directly.
	pass, follow, _ := r.Analyse("qa_lead", "1")
	fmt.Printf("QA Lead agrees? %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}

	_ = r.EvaluateGates([]string{"clarity-form-1", "duplicate-1"})
	for _, gr := range r.GateResults {
		fmt.Printf("Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
	}
}
