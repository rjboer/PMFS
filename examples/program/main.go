package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
		fmt.Println("3) Show project overview")
		fmt.Println("4) Exit")
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
			showOverview(prj)
		case "4", "exit":
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
