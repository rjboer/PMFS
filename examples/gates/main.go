package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
)

// This example demonstrates evaluating a requirement against a gate. Requires
// the GEMINI_API_KEY environment variable.
func main() {
	db, err := PMFS.LoadSetup(".")
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}
	req := PMFS.Requirement{Description: "The system shall be user friendly."}
	if err := req.EvaluateGates(db, []string{"clarity-form-1"}); err != nil {
		log.Fatalf("EvaluateGates: %v", err)
	}
	for _, gr := range req.GateResults {
		fmt.Printf("Gate %s passed: %v\n", gr.Gate.ID, gr.Pass)
	}
}
