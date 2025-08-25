package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
	llm "github.com/rjboer/PMFS/pmfs/llm"
)

// This example demonstrates a full flow using Gemini to analyze a document,
// store the returned requirement, ask a role-specific question about it, and
// evaluate it against quality gates. Requires the GEMINI_API_KEY environment
// variable.
func main() {
	PMFS.SetBaseDir(".")
	prj := PMFS.ProjectType{ProductID: 0, ID: 0, LLM: llm.DefaultClient}
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
	pass, follow, _ := r.Analyse(&prj, "qa_lead", "1")
	fmt.Printf("QA Lead agrees? %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}

	_ = r.EvaluateGates(&prj, []string{"clarity-form-1", "duplicate-1"})
	for _, gr := range r.GateResults {
		fmt.Printf("Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
	}
}
