package main

import (
	"fmt"
	"log"

	PMFS "github.com/rjboer/PMFS"
)

// This example demonstrates analyzing an attachment with a role-specific
// question. Requires the GEMINI_API_KEY environment variable.
func main() {
	_, err := PMFS.LoadSetup(".")
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}
	prj := PMFS.ProjectType{ProductID: 0, ID: 0}
	att := PMFS.Attachment{RelPath: "../../../testdata/spec1.txt"}

	pass, follow, err := att.AnalyzeWithRole("product_manager", "1", &prj)
	if err != nil {
		log.Fatalf("AnalyzeWithRole: %v", err)
	}
	fmt.Printf("Pass: %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}
}
