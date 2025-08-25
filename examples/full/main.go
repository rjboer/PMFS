package main

import (
	"fmt"
	"log"
	"strconv"

	PMFS "github.com/rjboer/PMFS"
	llm "github.com/rjboer/PMFS/pmfs/llm"
)

// This example demonstrates a full flow using Gemini to analyze a document,
// storing the returned requirements, asking multiple role-specific questions
// about each requirement, and evaluating them against quality gates. Requires
// the GEMINI_API_KEY environment variable.
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

	roles := []string{"product_manager", "qa_lead", "security_privacy_officer"}

	// Ask each role about every requirement and evaluate quality gates.
	for i := range prj.D.PotentialRequirements {
		r := &prj.D.PotentialRequirements[i]
		fmt.Printf("Requirement %d: %s - %s\n", i+1, r.Name, r.Description)
		id := strconv.Itoa(i + 1)
		for _, role := range roles {
			pass, follow, _ := r.Analyse(&prj, role, id)
			fmt.Printf("  %s agrees? %v\n", role, pass)
			if follow != "" {
				fmt.Printf("    Follow-up: %s\n", follow)
			}
		}

		_ = r.EvaluateGates(&prj, []string{"clarity-form-1", "duplicate-1"})
		for _, gr := range r.GateResults {
			fmt.Printf("  Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
		}
	}
}
