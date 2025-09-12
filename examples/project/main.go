package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	PMFS "github.com/rjboer/PMFS"
)

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// This example demonstrates setting up a project, ingesting an attachment,
// analyzing requirements and evaluating quality gates. Requires the
// GEMINI_API_KEY environment variable.
func main() {
	dir, err := os.MkdirTemp("", "pmfs-example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db, err := PMFS.LoadSetup(dir)
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}

	id, err := db.NewProduct(PMFS.ProductData{Name: "Demo Product"})
	if err != nil {
		log.Fatalf("NewProduct: %v", err)
	}
	p := &db.Products[id-1]
	prjID, err := p.NewProject(PMFS.ProjectData{Name: "Demo Project"})
	if err != nil {
		log.Fatalf("NewProject: %v", err)
	}
	prj, err := p.Project(prjID)
	if err != nil {
		log.Fatalf("Project: %v", err)
	}

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
	// Mark all suggested requirements as active so they can be processed.
	prj.ActivateRequirementsWhere(func(r PMFS.Requirement) bool { return true })

	for i := range prj.D.Requirements {
		r := &prj.D.Requirements[i]
		if !r.Condition.Active || r.Condition.Deleted {
			continue
		}
		pass, follow, err := r.Analyze("product_manager", "1")
		if err != nil {
			log.Fatalf("Requirement Analyze: %v", err)
		}
		fmt.Printf("%s agrees? %v\n", r.Name, pass)
		if follow != "" {
			fmt.Printf("  Follow-up: %s\n", follow)
		}
		if err := r.EvaluateGates([]string{"clarity-form-1"}); err != nil {
			log.Fatalf("EvaluateGates: %v", err)
		}
		for _, gr := range r.GateResults {
			fmt.Printf("  Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
		}
	}
}
