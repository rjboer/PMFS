package main

import (
	"fmt"
	"log"
	"strings"

	PMFS "github.com/rjboer/PMFS"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// This example demonstrates analysing an attachment with a role-specific question.
func main() {
	// Stub Gemini client so the example runs without external calls.
	stub := gemini.ClientFunc{
		AskFunc: func(prompt string) (string, error) {
			if strings.Contains(strings.ToLower(prompt), "answer yes or no only") {
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

	pass, follow, err := att.Analyse("product_manager", "1", &prj)
	if err != nil {
		log.Fatalf("Analyse: %v", err)
	}
	fmt.Printf("Pass: %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}
}
