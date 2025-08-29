package main

import (
	"log"
	"os"
	"time"

	PMFS "github.com/rjboer/PMFS"
)

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// This example demonstrates setting up a project, ingesting an attachment,
// analysing requirements and evaluating quality gates. Requires the
// GEMINI_API_KEY environment variable.
func main() {
	path := "./RoelofCompany"
	err := os.MkdirAll(path, 777)
	if err != nil {
		log.Fatal(err)
	}

	db, err := PMFS.LoadSetup(path)
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}
	//try to make a first product
	id, err := db.NewProduct(PMFS.ProductData{Name: "Demo Product"})
	if err != nil {
		log.Fatalf("NewProduct: %v", err)
	}
	//attach a pointer to the products
	p := &db.Products[id-1]
	prjID, err := p.NewProject(PMFS.ProjectData{Name: "Demo Project"})
	if err != nil {
		log.Fatalf("NewProject: %v", err)
	}

	prj, err := p.Project(prjID)
	prj.D.Priority = "high"
	prj.D.StartDate = time.Now()
	prj.D.EndDate = time.Now().Add(time.Hour * 24 * 10)

	x := PMFS.Requirement{}
	x.Name = "test"
	x.Description = "tester"
	prj.AddRequirement(x)
	PMFS.DB.Save()
	prj.ExportExcel("./test.xlsx")

}
