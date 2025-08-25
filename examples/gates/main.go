package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
	llm "github.com/rjboer/PMFS/pmfs/llm"
)

// This example demonstrates evaluating a requirement against a gate. Requires
// the GEMINI_API_KEY environment variable.
func main() {
	req := PMFS.Requirement{Description: "The system shall be user friendly."}
	prj := PMFS.Project{LLM: llm.DefaultClient}
	if err := req.EvaluateGates(&prj, []string{"clarity-form-1"}); err != nil {
		log.Fatalf("EvaluateGates: %v", err)
	}
	for _, gr := range req.GateResults {
		fmt.Printf("Gate %s passed: %v\n", gr.Gate.ID, gr.Pass)
	}
}
