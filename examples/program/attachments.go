package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	PMFS "github.com/rjboer/PMFS"
)

// copyFile copies the file from src to dst using standard permissions.
func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// ingestAttachment asks for a file path, copies it into the project's input directory, ingests the file, lets the user pick one attachment for analysis and saves any newly suggested requirements.
func ingestAttachment(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	fmt.Print("Path to attachment: ")
	if !scanner.Scan() {
		return
	}
	src := scanner.Text()

	inputDir := filepath.Join(PMFS.DB.BaseDir, "products", strconv.Itoa(prj.ProductID), "projects", strconv.Itoa(prj.ID), "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		log.Printf("create input dir: %v", err)
		return
	}
	dst := filepath.Join(inputDir, filepath.Base(src))
	if err := copyFile(src, dst); err != nil {
		log.Printf("copy file: %v", err)
		return
	}

	atts, err := prj.Attachments().AddFromInputFolder()
	if err != nil {
		log.Printf("AddFromInputFolder: %v", err)
		return
	}
	if len(atts) == 0 {
		fmt.Println("No attachments ingested.")
		return
	}

	fmt.Println("Ingested attachments:")
	for i, a := range atts {
		fmt.Printf("%d) %s\n", i+1, a.Filename)
	}
	fmt.Print("Select attachment to analyze: ")
	if !scanner.Scan() {
		return
	}
	idx, err := strconv.Atoi(scanner.Text())
	if err != nil || idx < 1 || idx > len(atts) {
		fmt.Println("Invalid selection")
		return
	}

	selectedID := atts[idx-1].ID
	var att *PMFS.Attachment
	for i := range prj.D.Attachments {
		if prj.D.Attachments[i].ID == selectedID {
			att = &prj.D.Attachments[i]
			break
		}
	}
	if att == nil {
		fmt.Println("Attachment not found")
		return
	}
	if att.Analyzed {
		fmt.Println("Attachment already analyzed.")
		return
	}

	before := len(prj.D.Requirements)
	if err := att.Analyze(prj); err != nil {
		log.Printf("Analyze: %v", err)
		return
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	newReqs := prj.D.Requirements[before:]
	if len(newReqs) == 0 {
		fmt.Println("No new requirements suggested.")
		return
	}
	fmt.Println("Newly suggested requirements:")
	for _, r := range newReqs {
		fmt.Printf("- %s: %s\n", r.Name, r.Description)
	}
}

// attachmentsMenu manages attachment ingestion.
func attachmentsMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s > Project: %s > Attachments", p.Name, prj.Name))
		options := []string{
			"Ingest attachment",
			"Back to project menu",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			ingestAttachment(scanner, prj)
		case 1:
			return
		case 2:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}
