package main

import (
	"fmt"
	"log"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// This example demonstrates using the Gemini client to analyze a document
// and answer a free-form question. It swaps in a stub client so the example
// runs without calling the real API.
func main() {
	prev := gemini.SetClient(gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return []gemini.Requirement{{ID: 1, Name: "Sample", Description: "From " + path}}, nil
		},
		AskFunc: func(prompt string) (string, error) {
			return "stubbed answer", nil
		},
	})
	defer gemini.SetClient(prev)

	reqs, err := gemini.AnalyzeAttachment("testdata/spec1.txt")
	if err != nil {
		log.Fatalf("analyze: %v", err)
	}
	for _, r := range reqs {
		fmt.Printf("Requirement: %s - %s\n", r.Name, r.Description)
	}

	ans, err := gemini.Ask("Summarize the spec")
	if err != nil {
		log.Fatalf("ask: %v", err)
	}
	fmt.Printf("Answer: %s\n", ans)
}
