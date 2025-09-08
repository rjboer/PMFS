package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	PMFS "github.com/rjboer/PMFS"
)

// loadEnv loads environment variables from a .env file if it exists.
func loadEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		os.Setenv(key, val)
	}
	if err := s.Err(); err != nil {
		log.Printf("Error reading .env: %v", err)
	}
}

// menuSelect displays a menu with the provided title and options and returns
// the selected index.
func menuSelect(title string, options []string) (int, error) {
	prompt := promptui.Select{Label: title, Items: options}
	idx, _, err := prompt.Run()
	return idx, err
}

// clearScreen clears the terminal using ANSI escape codes.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

var heading = color.New(color.FgCyan, color.Bold).SprintfFunc()

// listProducts prints all products in the database with their IDs.
func listProducts() {
	if len(PMFS.DB.Products) == 0 {
		fmt.Println("No products available.")
		return
	}
	fmt.Println("Products:")
	for _, p := range PMFS.DB.Products {
		fmt.Printf("%d: %s\n", p.ID, p.Name)
	}
}

// createProduct asks the user for a name, creates the product and saves the DB.
func createProduct(scanner *bufio.Scanner) *PMFS.ProductType {
	fmt.Print("Product name: ")
	if !scanner.Scan() {
		return nil
	}
	name := scanner.Text()
	id, err := PMFS.DB.NewProduct(PMFS.ProductData{Name: name})
	if err != nil {
		log.Printf("NewProduct: %v", err)
		return nil
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	p := &PMFS.DB.Products[id-1]
	fmt.Printf("Created product %s (ID: %d)\n", p.Name, p.ID)
	return p
}

// selectProduct displays a menu of product names and returns the chosen product.
func selectProduct(scanner *bufio.Scanner) *PMFS.ProductType {
	if len(PMFS.DB.Products) == 0 {
		fmt.Println("No products available.")
		return nil
	}

	options := make([]string, len(PMFS.DB.Products))
	for i, p := range PMFS.DB.Products {
		options[i] = p.Name
	}

	idx, err := menuSelect("Select product", options)
	if err != nil {
		return nil
	}
	return &PMFS.DB.Products[idx]
}

// selectProject prompts for a project ID and loads it from disk.
func selectProject(scanner *bufio.Scanner, p *PMFS.ProductType) *PMFS.ProjectType {
	if len(p.Projects) == 0 {
		fmt.Println("No projects available.")
		return nil
	}
	fmt.Println("Projects:")
	for _, pr := range p.Projects {
		fmt.Printf("%d: %s\n", pr.ID, pr.Name)
	}
	fmt.Print("Select project ID: ")
	if !scanner.Scan() {
		return nil
	}
	id, err := strconv.Atoi(scanner.Text())
	if err != nil {
		fmt.Println("Invalid selection")
		return nil
	}
	prj, err := p.Project(id)
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

// productMenu provides options for a selected product.
func productMenu(scanner *bufio.Scanner, p *PMFS.ProductType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s", p.Name))
		options := []string{
			"Project operations",
			"Edit product",
			"Back to product list",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			projectOpsMenu(scanner, p)
		case 1:
			fmt.Print("New product name: ")
			if !scanner.Scan() {
				return
			}
			newName := scanner.Text()
			if newName != "" {
				p.Name = newName
				if _, err := PMFS.DB.ModifyProduct(PMFS.ProductData{ID: p.ID, Name: newName}); err != nil {
					log.Printf("ModifyProduct: %v", err)
				} else if err := PMFS.DB.Save(); err != nil {
					log.Printf("Save DB: %v", err)
				}
			}
		case 2:
			return
		case 3:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
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
			if _, err := p.NewProject(PMFS.ProjectData{Name: name}); err != nil {
				log.Printf("NewProject: %v", err)
			} else if err := PMFS.DB.Save(); err != nil {
				log.Printf("Save DB: %v", err)
			}
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
// file at a user-specified path.
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

// importExcel loads project data from an Excel file and merges it into the
// current project or creates a new one when no project exists.
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

// copyFile copies the file from src to dst using standard permissions.
func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// legacy whiterows replaced by clearScreen

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

// showOverview prints current requirements and attachments of the project.
func showOverview(prj *PMFS.ProjectType) {
	fmt.Println("Requirements:")
	if len(prj.D.Requirements) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, r := range prj.D.Requirements {
			if r.Condition.Deleted {
				continue
			}
			state := "inactive"
			if r.Condition.Active {
				state = "active"
			} else if r.Condition.Proposed {
				state = "proposed"
			}
			fmt.Printf("  %d: %s - %s (%s)\n", r.ID, r.Name, r.Description, state)
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

// showRequirementsByStatus prints requirements filtered by status.
func showRequirementsByStatus(scanner *bufio.Scanner, prj *PMFS.ProjectType) {
	fmt.Print("Status (active, proposed, inactive): ")
	if !scanner.Scan() {
		return
	}
	status := strings.ToLower(strings.TrimSpace(scanner.Text()))
	fmt.Println("Requirements:")
	found := false
	for _, r := range prj.D.Requirements {
		if r.Condition.Deleted {
			continue
		}
		var match bool
		switch status {
		case "active":
			match = r.Condition.Active
		case "proposed":
			match = r.Condition.Proposed
		case "inactive":
			match = !r.Condition.Active && !r.Condition.Proposed
		default:
			fmt.Println("Unknown status")
			return
		}
		if match {
			fmt.Printf("  %d: %s - %s\n", r.ID, r.Name, r.Description)
			found = true
		}
	}
	if !found {
		fmt.Println("  (none)")
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
		if r.Condition.Deleted {
			continue
		}
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
	if !req.Condition.Active || req.Condition.Deleted {
		fmt.Println("Requirement is not active.")
		return
	}
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
		if r.Condition.Deleted {
			continue
		}
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
	if req.Condition.Deleted {
		fmt.Println("Requirement is deleted.")
		return
	}
	others, err := req.SuggestOthers(prj)
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

// generateDesignAspects runs GenerateDesignAspectsAll on the project.
func generateDesignAspects(prj *PMFS.ProjectType) {
	if err := prj.GenerateDesignAspectsAll(); err != nil {
		log.Printf("GenerateDesignAspectsAll: %v", err)
		return
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	fmt.Println("Design aspects generated.")
}

// generateRequirementsFromAspects creates requirements from stored design aspects.
func generateRequirementsFromAspects(prj *PMFS.ProjectType) {
	added := 0
	for i := range prj.D.Requirements {
		for j := range prj.D.Requirements[i].DesignAspects {
			da := &prj.D.Requirements[i].DesignAspects[j]
			if da.Processed {
				continue
			}
			reqs, err := da.GenerateTemplates("product_manager", "1")
			if err != nil {
				log.Printf("GenerateTemplates: %v", err)
				continue
			}
			if err := da.EvaluateDesignGates(PMFS.DesignAspectGateGroup); err != nil {
				log.Printf("EvaluateDesignGates: %v", err)
			}
			for k := range reqs {
				reqs[k].ID = len(prj.D.Requirements) + 1
				reqs[k].Condition.Proposed = true
				reqs[k].Condition.AIgenerated = true
				prj.D.Requirements = append(prj.D.Requirements, reqs[k])
				added++
			}
			da.Processed = true
		}
	}
	if added > 0 {
		prj.D.Requirements = PMFS.Deduplicate(prj.D.Requirements, false)
		if err := prj.Save(); err != nil {
			log.Printf("Save project: %v", err)
		} else if err := PMFS.DB.Save(); err != nil {
			log.Printf("Save DB: %v", err)
		}
	}
	fmt.Printf("Generated %d requirements from design aspects.\n", added)
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

// requirementsMenu manages requirement operations.
func requirementsMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s > Project: %s > Requirements", p.Name, prj.Name))
		options := []string{
			"Add requirement",
			"Show project overview",
			"Show requirements by status",
			"Generate design aspects",
			"Generate requirements from design aspects",
			"Back to project menu",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			addRequirement(scanner, prj)
		case 1:
			showOverview(prj)
		case 2:
			showRequirementsByStatus(scanner, prj)
		case 3:
			generateDesignAspects(prj)
		case 4:
			generateRequirementsFromAspects(prj)
		case 5:
			return
		case 6:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
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

// This example demonstrates basic project interaction via a simple command
// loop. It allows adding requirements, importing/exporting to Excel and viewing
// the project's current state. The program reads GEMINI_API_KEY from the
// environment and, if present, loads it from a .env file in the working
// directory.
func main() {
	loadEnv()
	if _, ok := os.LookupEnv("GEMINI_API_KEY"); !ok {
		fmt.Fprintln(os.Stderr, "GEMINI_API_KEY environment variable not set")
		fmt.Fprintln(os.Stderr, "Press any key to continue...")
		_, _ = bufio.NewReader(os.Stdin).ReadByte()
		os.Exit(1)
	}

	path := "./RoelofCompany"
	if err := os.MkdirAll(path, 0o777); err != nil {
		log.Fatal(err)
	}

	if _, err := PMFS.LoadSetup(path); err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		clearScreen()
		listProducts()
		options := []string{"Select product", "Create product", "Exit"}
		idx, err := menuSelect("Product menu", options)
		if err != nil {
			break
		}
		switch idx {
		case 0:
			p := selectProduct(scanner)
			if p != nil {
				productMenu(scanner, p)
			}
		case 1:
			createProduct(scanner)
		case 2:
			fmt.Println("Goodbye!")
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("input error: %v", err)
	}
}
