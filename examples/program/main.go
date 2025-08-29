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
// analysing requirements and evaluating quality gates. Requires the
// GEMINI_API_KEY environment variable.
func main() {
	path := "./RoelofCompany"
	err := os.MkdirAll(path, 777)
	if err != nil {
		log.Fatal(err)
	}

	db, err := PMFS.LoadSetup(path)
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}
	//try to make a first product
	id, err := db.NewProduct(PMFS.ProductData{Name: "Demo Product"})
	if err != nil {
		log.Fatalf("NewProduct: %v", err)
	}
	//attach a pointer to the products
	p := &db.Products[id-1]
	prjID, err := p.NewProject(db, PMFS.ProjectData{Name: "Demo Project"})
	if err != nil {
		log.Fatalf("NewProject: %v", err)
	}

	prj, err := p.Project(prjID)
	prj.D.Priority = "high"
	prj.D.StartDate = time.Now()
	prj.D.EndDate = time.Now().Add(time.Hour * 24 * 10)

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

	if err := prj.D.Attachments[0].Analyze(db, prj); err != nil {
		log.Fatalf("Attachment Analyze: %v", err)
	}

	for i := range prj.D.PotentialRequirements {
		r := &prj.D.PotentialRequirements[i]
		pass, follow, err := r.Analyse(db, "product_manager", "1")
		if err != nil {
			log.Fatalf("Requirement Analyse: %v", err)
		}
		fmt.Printf("%s agrees? %v\n", r.Name, pass)
		if follow != "" {
			fmt.Printf("  Follow-up: %s\n", follow)
		}
		if err := r.EvaluateGates(db, []string{"clarity-form-1"}); err != nil {
			log.Fatalf("EvaluateGates: %v", err)
		}
		for _, gr := range r.GateResults {
			fmt.Printf("  Gate %s passed? %v\n", gr.Gate.ID, gr.Pass)
		}
	}
}
