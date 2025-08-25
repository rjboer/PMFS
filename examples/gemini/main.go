package main

import (
	"fmt"
	"log"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// This example demonstrates using the Gemini client to analyze a document and
// answer a free-form question. Requires the GEMINI_API_KEY environment
// variable.
func main() {
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
