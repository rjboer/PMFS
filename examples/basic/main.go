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
	if err := PMFS.EnsureLayout(); err != nil {
		log.Fatalf("ensure layout: %v", err)
	}

	idx, err := PMFS.LoadIndex()
	if err != nil {
		log.Fatalf("load index: %v", err)
	}

	if len(idx.Products) == 0 {
		if err := idx.AddProduct("Example Product"); err != nil {
			log.Fatalf("add product: %v", err)
		}
		if err := idx.Products[0].AddProject(&idx, "Example Project"); err != nil {
			log.Fatalf("add project: %v", err)
		}
	}

	for _, p := range idx.Products {
		fmt.Printf("Product %d: %s\n", p.ID, p.Name)
		for _, pr := range p.Projects {
			fmt.Printf("  Project %d: %s\n", pr.ID, pr.Name)
		}
	}
}
