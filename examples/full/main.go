package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	PMFS "github.com/rjboer/PMFS"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// This example demonstrates a full flow using Gemini to analyze a document,
// storing the returned requirements, asking multiple role-specific questions
// about each requirement, and evaluating them against quality gates. After the
// Gemini client is configured, requirement methods can be used directly without
// passing the client to interact or gates packages. Remove the stub below and
// set GEMINI_API_KEY to call the real API with the default client.
func main() {
	// Stub the Gemini client so the example runs without external calls.
	// Delete this block for live API calls.
	stub := gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return []gemini.Requirement{
				{ID: 1, Name: "Login", Description: "Users shall log in with email and password."},
				{ID: 2, Name: "Logout", Description: "Users shall be able to log out securely."},
			}, nil
		},
		AskFunc: func(prompt string) (string, error) {
			p := strings.ToLower(prompt)
			if strings.Contains(p, "answer yes or no only") {
				return "Yes", nil
			}
			if strings.Contains(p, "given the requirement") {
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

	roles := []string{"product_manager", "qa_lead", "security_privacy_officer"}

	// Ask each role about every requirement and evaluate quality gates.
	for i := range prj.D.PotentialRequirements {
		r := &prj.D.PotentialRequirements[i]
		fmt.Printf("Requirement %d: %s - %s\n", i+1, r.Name, r.Description)
		id := strconv.Itoa(i + 1)
		for _, role := range roles {
			pass, follow, _ := r.Analyse(role, id)
			fmt.Printf("  %s agrees? %v\n", role, pass)
			if follow != "" {
				fmt.Printf("    Follow-up: %s\n", follow)
			}
		}

		_ = r.EvaluateGates([]string{"clarity-form-1", "duplicate-1"})
		for _, gr := range r.GateResults {
			fmt.Printf("  Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
		}
	}
}
