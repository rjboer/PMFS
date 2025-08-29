package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	PMFS "github.com/rjboer/PMFS"
)

// addRequirement prompts the user for requirement details and appends it to
// the project.
func addRequirement(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	fmt.Print("Requirement name: ")
	if !scanner.Scan() {
		return
	}
	name := scanner.Text()

	fmt.Print("Description: ")
	if !scanner.Scan() {
		return
	}
	desc := scanner.Text()

	r := PMFS.Requirement{Name: name, Description: desc}
	if err := prj.AddRequirement(r); err != nil {
		log.Printf("AddRequirement: %v", err)
		return
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
}

// exportExcel saves the database and writes the project overview to an Excel
// file.
func exportExcel(prj *PMFS.ProjectType) {
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	if err := prj.ExportExcel("./test.xlsx"); err != nil {
		log.Printf("ExportExcel: %v", err)
	}
}

// copyFile copies the file from src to dst using standard permissions.
func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// ingestAttachment asks for a file path, copies it into the project's input
// directory, ingests the file, lets the user pick one attachment for analysis
// and saves any newly suggested requirements.
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

	before := len(prj.D.PotentialRequirements)
	if err := att.Analyze(prj); err != nil {
		log.Printf("Analyze: %v", err)
		return
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	newReqs := prj.D.PotentialRequirements[before:]
	if len(newReqs) == 0 {
		fmt.Println("No new requirements suggested.")
		return
	}
	fmt.Println("Newly suggested requirements:")
	for _, r := range newReqs {
		fmt.Printf("- %s: %s\n", r.Name, r.Description)
	}
}

// showOverview prints current requirements and attachments of the project.
func showOverview(prj *PMFS.ProjectType) {
	fmt.Println("Requirements:")
	if len(prj.D.Requirements) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, r := range prj.D.Requirements {
			fmt.Printf("  %d: %s - %s\n", r.ID, r.Name, r.Description)
		}
	}

	fmt.Println("Attachments:")
	if len(prj.D.Attachments) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, a := range prj.D.Attachments {
			fmt.Printf("  %d: %s\n", a.ID, a.Filename)
		}
	}
}

// analyseRequirement lists existing requirements, lets the user choose one and
// runs QualityControlAI on it, printing the results.
func analyseRequirement(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	if len(prj.D.Requirements) == 0 {
		fmt.Println("No requirements to analyse.")
		return
	}
	fmt.Println("Select requirement to analyse:")
	for i, r := range prj.D.Requirements {
		fmt.Printf("%d) %s - %s\n", i+1, r.Name, r.Description)
	}
	fmt.Print("> ")
	if !scanner.Scan() {
		return
	}
	idx, err := strconv.Atoi(scanner.Text())
	if err != nil || idx < 1 || idx > len(prj.D.Requirements) {
		fmt.Println("Invalid selection")
		return
	}
	req := &prj.D.Requirements[idx-1]
	pass, follow, err := req.QualityControlAI("product_manager", "1", []string{"clarity-form-1"})
	if err != nil {
		log.Printf("QualityControlAI: %v", err)
		return
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	fmt.Printf("Analysis pass: %v\n", pass)
	if follow != "" {
		fmt.Printf("Follow-up: %s\n", follow)
	}
	fmt.Println("Gate results:")
	for _, gr := range req.GateResults {
		fmt.Printf("  %s: %v\n", gr.Gate.ID, gr.Pass)
	}
}

// suggestRelated lets the user pick a requirement and asks the LLM for related
// requirements, printing any suggestions.
func suggestRelated(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	if len(prj.D.Requirements) == 0 {
		fmt.Println("No requirements available.")
		return
	}
	fmt.Println("Select requirement for suggestions:")
	for i, r := range prj.D.Requirements {
		fmt.Printf("%d) %s - %s\n", i+1, r.Name, r.Description)
	}
	fmt.Print("> ")
	if !scanner.Scan() {
		return
	}
	idx, err := strconv.Atoi(scanner.Text())
	if err != nil || idx < 1 || idx > len(prj.D.Requirements) {
		fmt.Println("Invalid selection")
		return
	}
	req := &prj.D.Requirements[idx-1]
	others, err := req.SuggestOthers()
	if err != nil {
		log.Printf("SuggestOthers: %v", err)
		return
	}
	if len(others) == 0 {
		fmt.Println("No related requirements suggested.")
		return
	}
	fmt.Println("Suggested requirements:")
	for _, r := range others {
		fmt.Printf("- %s: %s\n", r.Name, r.Description)
	}
}

// This example demonstrates basic project interaction via a simple command
// loop. It allows adding requirements, exporting to Excel and viewing the
// project's current state. Requires the GEMINI_API_KEY environment variable.
func main() {
	path := "./RoelofCompany"
	if err := os.MkdirAll(path, 0o777); err != nil {
		log.Fatal(err)
	}

	db, err := PMFS.LoadSetup(path)
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

	prj.D.Priority = "high"
	prj.D.StartDate = time.Now()
	prj.D.EndDate = time.Now().Add(time.Hour * 24 * 10)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Choose an option:")
		fmt.Println("1) Add requirement")
		fmt.Println("2) Export to Excel")
		fmt.Println("3) Ingest attachment")
		fmt.Println("4) Show project overview")
		fmt.Println("5) Analyse requirement")
		fmt.Println("6) Suggest related requirements")
		fmt.Println("7) Exit")
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			addRequirement(scanner, prj)
		case "2":
			exportExcel(prj)
		case "3":
			ingestAttachment(scanner, prj)
		case "4":
			showOverview(prj)
		case "5":
			analyseRequirement(scanner, prj)
		case "6":
			suggestRelated(scanner, prj)
		case "7", "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Unknown option")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("input error: %v", err)
	}
}
