package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	PMFS "github.com/rjboer/PMFS"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// This example demonstrates setting up a project, ingesting an attachment,
// analysing requirements and evaluating quality gates.
func main() {
	stub := gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return []gemini.Requirement{
				{ID: 1, Name: "Register", Description: "Users shall register with email."},
				{ID: 2, Name: "Confirm", Description: "System shall send confirmation email."},
			}, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if strings.Contains(strings.ToLower(prompt), "answer yes or no") {
				return "Yes", nil
			}
			if strings.Contains(strings.ToLower(prompt), "follow-up") {
				return "details", nil
			}
			if strings.Contains(strings.ToLower(prompt), "given the requirement") {
				return "Yes", nil
			}
			return "Yes", nil
		},
	}
	prev := gemini.SetClient(stub)
	defer gemini.SetClient(prev)

	dir, err := os.MkdirTemp("", "pmfs-example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	PMFS.SetBaseDir(dir)
	if err := PMFS.EnsureLayout(); err != nil {
		log.Fatalf("EnsureLayout: %v", err)
	}

	idx, err := PMFS.LoadIndex()
	if err != nil {
		log.Fatalf("LoadIndex: %v", err)
	}
	if err := idx.AddProduct("Demo Product"); err != nil {
		log.Fatalf("AddProduct: %v", err)
	}
	if err := idx.Products[0].AddProject(&idx, "Demo Project"); err != nil {
		log.Fatalf("AddProject: %v", err)
	}
	prj := &idx.Products[0].Projects[0]

	attDir := filepath.Join(dir, "products", "1", "projects", "1", "attachments", "1")
	if err := os.MkdirAll(attDir, 0o755); err != nil {
		log.Fatalf("mkdir attDir: %v", err)
	}
	src := filepath.Join("testdata", "spec1.txt")
	dst := filepath.Join(attDir, "spec1.txt")
	if err := copyFile(src, dst); err != nil {
		log.Fatalf("copy file: %v", err)
	}

	att := PMFS.Attachment{
		ID:       1,
		Filename: "spec1.txt",
		RelPath:  filepath.ToSlash(filepath.Join("attachments", "1", "spec1.txt")),
		Mimetype: "text/plain",
		AddedAt:  time.Now(),
	}
	prj.D.Attachments = append(prj.D.Attachments, att)

	if err := prj.D.Attachments[0].Analyze(prj); err != nil {
		log.Fatalf("Attachment Analyze: %v", err)
	}

	for i := range prj.D.PotentialRequirements {
		r := &prj.D.PotentialRequirements[i]
		pass, follow, err := r.Analyse(prj, "product_manager", "1")
		if err != nil {
			log.Fatalf("Requirement Analyse: %v", err)
		}
		fmt.Printf("%s agrees? %v\n", r.Name, pass)
		if follow != "" {
			fmt.Printf("  Follow-up: %s\n", follow)
		}
		if err := r.EvaluateGates(prj, []string{"clarity-form-1"}); err != nil {
			log.Fatalf("EvaluateGates: %v", err)
		}
		for _, gr := range r.GateResults {
			fmt.Printf("  Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
		}
	}
}
