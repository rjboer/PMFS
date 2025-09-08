package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	PMFS "github.com/rjboer/PMFS"
)

// addRequirement prompts the user for requirement details and appends it to the project.
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

// analyseRequirement lists existing requirements, lets the user choose one and runs QualityControlAI on it, printing the results.
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

// suggestRelated lets the user pick a requirement and asks the LLM for related requirements, printing any suggestions.
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
