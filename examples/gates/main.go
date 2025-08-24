package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// This example demonstrates evaluating a requirement against a gate.
func main() {
	// Stub Gemini client so the example runs without external calls.
	stub := gemini.ClientFunc{
		AskFunc: func(prompt string) (string, error) {
			return "Yes", nil
		},
	}
	prev := gemini.SetClient(stub)
	defer gemini.SetClient(prev)

	req := PMFS.Requirement{Description: "The system shall be user friendly."}
	if err := req.EvaluateGates([]string{"clarity-form-1"}); err != nil {
		log.Fatalf("EvaluateGates: %v", err)
	}
	for _, gr := range req.GateResults {
		fmt.Printf("Gate %s passed: %v\n", gr.Gate.ID, gr.Pass)
	}
}
