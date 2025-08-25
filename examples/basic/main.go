package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
)

// This example demonstrates basic usage of the PMFS package.
// It ensures the directory layout exists, adds a product and project
// if none are present, and then prints the structure.
func main() {
	db, err := PMFS.LoadSetup("database")
	if err != nil {
		log.Fatalf("setup: %v", err)
	}

	if len(db.Index.Products) == 0 {
		p, err := db.AddProduct("Example Product")
		if err != nil {
			log.Fatalf("add product: %v", err)
		}
		if _, err := p.AddProject("Example Project"); err != nil {
			log.Fatalf("add project: %v", err)
		}
		if err := db.Save(); err != nil {
			log.Fatalf("save index: %v", err)
		}
	}

	for _, p := range db.Index.Products {
		fmt.Printf("Product %d: %s\n", p.ID, p.Name)
		for _, pr := range p.Projects {
			fmt.Printf("  Project %d: %s\n", pr.ID, pr.Name)
		}
	}
}
