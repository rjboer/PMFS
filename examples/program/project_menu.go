package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"

	PMFS "github.com/rjboer/PMFS"
)

// selectProject displays a menu of project names and returns the chosen project.
func selectProject(_ *bufio.Scanner, p *PMFS.ProductType) *PMFS.ProjectType {
	if len(p.Projects) == 0 {
		fmt.Println("No projects available.")
		return nil
	}

	options := make([]string, len(p.Projects))
	for i, pr := range p.Projects {
		options[i] = pr.Name
	}

	prompt := promptui.Select{Label: "Select project", Items: options}
	idx, _, err := prompt.Run()
	if err != nil {
		return nil
	}

	prj, err := p.Project(p.Projects[idx].ID)
	if err != nil {
		fmt.Printf("Load project: %v\n", err)
		return nil
	}
	return prj
}

// editProject interactively updates project fields and persists them.
func editProject(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	fmt.Printf("Name [%s]: ", prj.D.Name)
	if scanner.Scan() {
		if txt := scanner.Text(); txt != "" {
			prj.Name = txt
			prj.D.Name = txt
		}
	}
	fmt.Printf("Scope [%s]: ", prj.D.Scope)
	if scanner.Scan() {
		if txt := scanner.Text(); txt != "" {
			prj.D.Scope = txt
		}
	}
	fmt.Printf("Priority [%s]: ", prj.D.Priority)
	if scanner.Scan() {
		if txt := scanner.Text(); txt != "" {
			prj.D.Priority = txt
		}
	}
	fmt.Printf("Start date [%s] (YYYY-MM-DD): ", prj.D.StartDate.Format("2006-01-02"))
	if scanner.Scan() {
		if txt := scanner.Text(); txt != "" {
			if t, err := time.Parse("2006-01-02", txt); err == nil {
				prj.D.StartDate = t
			} else {
				fmt.Println("Invalid date format")
			}
		}
	}
	fmt.Printf("End date [%s] (YYYY-MM-DD): ", prj.D.EndDate.Format("2006-01-02"))
	if scanner.Scan() {
		if txt := scanner.Text(); txt != "" {
			if t, err := time.Parse("2006-01-02", txt); err == nil {
				prj.D.EndDate = t
			} else {
				fmt.Println("Invalid date format")
			}
		}
	}
	if err := prj.Save(); err != nil {
		log.Printf("Save project: %v", err)
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
}

// projectOpsMenu handles project-related operations for a product.
func projectOpsMenu(scanner *bufio.Scanner, p *PMFS.ProductType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s > Project operations", p.Name))
		options := []string{
			"List projects",
			"Create project",
			"Edit project",
			"Delete project",
			"Select project",
			"Back to product menu",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			if len(p.Projects) == 0 {
				fmt.Println("No projects available.")
			} else {
				for _, pr := range p.Projects {
					fmt.Printf("%d: %s\n", pr.ID, pr.Name)
				}
			}
		case 1:
			fmt.Print("Project name: ")
			if !scanner.Scan() {
				return
			}
			name := scanner.Text()
			id, err := p.NewProject(PMFS.ProjectData{Name: name})
			if err != nil {
				log.Printf("NewProject: %v", err)
				continue
			}
			if err := PMFS.DB.Save(); err != nil {
				log.Printf("Save DB: %v", err)
			}
			prj, err := p.Project(id)
			if err != nil {
				log.Printf("Load project: %v", err)
				continue
			}
			projectMenu(scanner, p, prj)
		case 2:
			prj := selectProject(scanner, p)
			if prj != nil {
				editProject(scanner, prj)
			}
		case 3:
			prj := selectProject(scanner, p)
			if prj != nil {
				prjDir := filepath.Join(PMFS.DB.BaseDir, "products", strconv.Itoa(p.ID), "projects", strconv.Itoa(prj.ID))
				if err := os.RemoveAll(prjDir); err != nil {
					log.Printf("Remove project: %v", err)
				} else {
					for i := range p.Projects {
						if p.Projects[i].ID == prj.ID {
							p.Projects = append(p.Projects[:i], p.Projects[i+1:]...)
							break
						}
					}
					if err := PMFS.DB.Save(); err != nil {
						log.Printf("Save DB: %v", err)
					}
				}
			}
		case 4:
			prj := selectProject(scanner, p)
			if prj != nil {
				projectMenu(scanner, p, prj)
			}
		case 5:
			return
		case 6:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}

// exportExcel saves the database and writes the project overview to an Excel file at a user-specified path.
func exportExcel(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	fmt.Print("Output path: ")
	if !scanner.Scan() {
		return
	}
	path := scanner.Text()

	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	if err := prj.ExportExcel(path); err != nil {
		fmt.Printf("Export failed: %v\n", err)
		return
	}
	fmt.Printf("Project exported to %s\n", path)
}

// exportProjectStruct writes the full project struct to a JSON file.
func exportProjectStruct(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	fmt.Print("Output path: ")
	if !scanner.Scan() {
		return
	}
	path := scanner.Text()

	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	data, err := json.MarshalIndent(prj, "", "  ")
	if err != nil {
		fmt.Printf("Export failed: %v\n", err)
		return
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Printf("Export failed: %v\n", err)
		return
	}
	fmt.Printf("Project struct exported to %s\n", path)
}

// importExcel loads project data from an Excel file and merges it into the current project or creates a new one when no project exists.
func importExcel(scanner *bufio.Scanner, p *PMFS.ProductType, prj **PMFS.ProjectType) {
	fmt.Print("Path to Excel file: ")
	if !scanner.Scan() {
		return
	}
	path := scanner.Text()

	data, err := PMFS.ImportProjectExcel(path)
	if err != nil {
		fmt.Printf("Import failed: %v\n", err)
		return
	}

	if *prj == nil {
		id, err := p.NewProject(*data)
		if err != nil {
			fmt.Printf("Create project: %v\n", err)
			return
		}
		np, err := p.Project(id)
		if err != nil {
			fmt.Printf("Load project: %v\n", err)
			return
		}
		*prj = np
		fmt.Println("Created new project from Excel data.")
	} else {
		(*prj).Name = data.Name
		(*prj).D.Name = data.Name
		(*prj).D.Scope = data.Scope
		(*prj).D.StartDate = data.StartDate
		(*prj).D.EndDate = data.EndDate
		(*prj).D.Status = data.Status
		(*prj).D.Priority = data.Priority
		(*prj).D.Requirements = append((*prj).D.Requirements, data.Requirements...)
		(*prj).D.Intelligence = append((*prj).D.Intelligence, data.Intelligence...)
		fmt.Println("Merged Excel data into current project.")
	}

	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
}

// analysisMenu groups requirement analysis operations.
func analysisMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s > Project: %s > Analysis", p.Name, prj.Name))
		options := []string{
			"Analyse requirement",
			"Suggest related requirements",
			"Back to project menu",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			analyseRequirement(scanner, prj)
		case 1:
			suggestRelated(scanner, prj)
		case 2:
			return
		case 3:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}

// exportImportMenu handles exporting and importing project data.
func exportImportMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj **PMFS.ProjectType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s > Project: %s > Export/Import", p.Name, (*prj).Name))
		options := []string{
			"Export to Excel",
			"Export project struct",
			"Import from Excel",
			"Back to project menu",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			exportExcel(scanner, *prj)
		case 1:
			exportProjectStruct(scanner, *prj)
		case 2:
			importExcel(scanner, p, prj)
		case 3:
			return
		case 4:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}

// projectMenu handles project-specific operations for a loaded project.
func projectMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s > Project: %s", p.Name, prj.Name))
		options := []string{
			"Requirements",
			"Attachments",
			"Analysis",
			"Export/Import",
			"Back to product menu",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			requirementsMenu(scanner, p, prj)
		case 1:
			attachmentsMenu(scanner, p, prj)
		case 2:
			analysisMenu(scanner, p, prj)
		case 3:
			exportImportMenu(scanner, p, &prj)
		case 4:
			return
		case 5:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}
