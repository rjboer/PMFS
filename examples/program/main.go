package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// selectProduct prompts for a product ID and returns the matching product.
func selectProduct(scanner *bufio.Scanner) *PMFS.ProductType {
	if len(PMFS.DB.Products) == 0 {
		fmt.Println("No products available.")
		return nil
	}
	listProducts()
	fmt.Print("Select product ID: ")
	if !scanner.Scan() {
		return nil
	}
	id, err := strconv.Atoi(scanner.Text())
	if err != nil {
		fmt.Println("Invalid ID")
		return nil
	}
	for i := range PMFS.DB.Products {
		if PMFS.DB.Products[i].ID == id {
			return &PMFS.DB.Products[i]
		}
	}
	fmt.Println("Product not found")
	return nil
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
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s\n", p.Name)
		fmt.Println()
		fmt.Println("1) Project operations")
		fmt.Println("2) Edit product")
		fmt.Println("3) Back to product list")
		fmt.Println("99) Exit")
		fmt.Println("-----------------------------------------")
		fmt.Println()
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			projectOpsMenu(scanner, p)
		case "2":
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
		case "3", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
		}
	}
}

// projectOpsMenu handles project-related operations for a product.
func projectOpsMenu(scanner *bufio.Scanner, p *PMFS.ProductType) {
	for {
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s > Project operations\n", p.Name)
		fmt.Println()
		fmt.Println("1) List projects")
		fmt.Println("2) Create project")
		fmt.Println("3) Edit project")
		fmt.Println("4) Delete project")
		fmt.Println("5) Select project")
		fmt.Println("6) Back to product menu")
		fmt.Println("99) Exit")
		fmt.Println("-----------------------------------------")
		fmt.Println()
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			if len(p.Projects) == 0 {
				fmt.Println("No projects available.")
			} else {
				for _, pr := range p.Projects {
					fmt.Printf("%d: %s\n", pr.ID, pr.Name)
				}
			}
		case "2":
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
		case "3":
			prj := selectProject(scanner, p)
			if prj != nil {
				editProject(scanner, prj)
			}
		case "4":
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
		case "5":
			prj := selectProject(scanner, p)
			if prj != nil {
				projectMenu(scanner, p, prj)
			}
		case "6", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
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
		(*prj).D.PotentialRequirements = append((*prj).D.PotentialRequirements, data.PotentialRequirements...)
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

func whiterows() {
	fmt.Printf("\n\n\n\n\n\n\n\n\n")
	fmt.Println("--------------------------------------------------------")
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

// projectMenu handles project-specific operations for a loaded project.
func projectMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s > Project: %s\n", p.Name, prj.Name)
		fmt.Println()
		fmt.Println("1) Requirements")
		fmt.Println("2) Attachments")
		fmt.Println("3) Analysis")
		fmt.Println("4) Export/Import")
		fmt.Println("5) Back to product menu")
		fmt.Println("99) Exit")
		fmt.Println()
		fmt.Println("--------------------------------------------------------")
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			requirementsMenu(scanner, p, prj)
		case "2":
			attachmentsMenu(scanner, p, prj)
		case "3":
			analysisMenu(scanner, p, prj)
		case "4":
			exportImportMenu(scanner, p, &prj)
		case "5", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
		}
	}
}

// requirementsMenu manages requirement operations.
func requirementsMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s > Project: %s > Requirements\n", p.Name, prj.Name)
		fmt.Println()
		fmt.Println("1) Add requirement")
		fmt.Println("2) Show project overview")
		fmt.Println("3) Back to project menu")
		fmt.Println("99) Exit")
		fmt.Println("--------------------------------------------------------")
		fmt.Println()
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			addRequirement(scanner, prj)
		case "2":
			showOverview(prj)
		case "3", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
		}
	}
}

// attachmentsMenu manages attachment ingestion.
func attachmentsMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s > Project: %s > Attachments\n", p.Name, prj.Name)
		fmt.Println()
		fmt.Println("1) Ingest attachment")
		fmt.Println("2) Back to project menu")
		fmt.Println("99) Exit")
		fmt.Println("--------------------------------------------------------")
		fmt.Println()
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			ingestAttachment(scanner, prj)
		case "2", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
		}
	}
}

// analysisMenu groups requirement analysis operations.
func analysisMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj *PMFS.ProjectType) {
	for {
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s > Project: %s > Analysis\n", p.Name, prj.Name)
		fmt.Println()
		fmt.Println("1) Analyse requirement")
		fmt.Println("2) Suggest related requirements")
		fmt.Println("3) Back to project menu")
		fmt.Println("99) Exit")
		fmt.Println("--------------------------------------------------------")
		fmt.Println()
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			analyseRequirement(scanner, prj)
		case "2":
			suggestRelated(scanner, prj)
		case "3", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
		}
	}
}

// exportImportMenu handles Excel export and import.
func exportImportMenu(scanner *bufio.Scanner, p *PMFS.ProductType, prj **PMFS.ProjectType) {
	for {
		fmt.Println()
		whiterows()
		fmt.Printf("Product: %s > Project: %s > Export/Import\n", p.Name, (*prj).Name)
		fmt.Println()
		fmt.Println("1) Export to Excel")
		fmt.Println("2) Import from Excel")
		fmt.Println("3) Back to project menu")
		fmt.Println("99) Exit")
		fmt.Println("--------------------------------------------------------")
		fmt.Println()
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			exportExcel(scanner, *prj)
		case "2":
			importExcel(scanner, p, prj)
		case "3", "back":
			return
		case "99", "exit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown option")
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
		whiterows()
		listProducts()
		fmt.Println()
		fmt.Println("Product menu:")
		fmt.Println()
		fmt.Println("1) Select product")
		fmt.Println("2) Create product")
		fmt.Println("3) Exit")
		fmt.Println("--------------------------------------------------------")
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			p := selectProduct(scanner)
			if p != nil {
				productMenu(scanner, p)
			}
		case "2":
			createProduct(scanner)
		case "3", "exit":
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
