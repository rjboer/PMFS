package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
)

// This example demonstrates a full flow using Gemini to analyze a document,
// store the returned requirement, ask a role-specific question about it, and
// evaluate it against quality gates. Requires the GEMINI_API_KEY environment
// variable.
func main() {

	_, err := PMFS.LoadSetup(".")
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}

	prj := PMFS.ProjectType{ProductID: 0, ID: 0}
	att := PMFS.Attachment{RelPath: "../../../testdata/spec1.txt"}

	// Analyze a document to extract requirements.
	if err := att.Analyze(&prj); err != nil {
		log.Fatalf("analyze: %v", err)
	}
	if len(prj.D.Requirements) == 0 {
		log.Fatal("no requirements returned")
	}
	// Activate all suggested requirements so we can operate on active ones.
	prj.ActivateRequirementsWhere(func(r PMFS.Requirement) bool { return true })

	var r *PMFS.Requirement
	for i := range prj.D.Requirements {
		if prj.D.Requirements[i].Condition.Active && !prj.D.Requirements[i].Condition.Deleted {
			r = &prj.D.Requirements[i]
			break
		}
	}
	if r == nil {
		log.Fatal("no active requirements")
	}

	fmt.Printf("Requirement: %s - %s\n", r.Name, r.Description)

	// With the client configured above, the requirement can query roles and
	// evaluate gates directly.
	pass, follow, _ := r.Analyze("qa_lead", "1")
	fmt.Printf("QA Lead agrees? %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}

	_ = r.EvaluateGates([]string{"clarity-form-1", "duplicate-1"})
	for _, gr := range r.GateResults {
		fmt.Printf("Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
	}
}
