package PMFS

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestExportPotentialRequirementsEnableColumn(t *testing.T) {
	prj := &ProjectType{D: ProjectData{PotentialRequirements: []Requirement{{
		ID:        1,
		Name:      "Pot",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}}

	tmp, err := os.CreateTemp("", "proj-*.xlsx")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := prj.ExportExcel(tmp.Name()); err != nil {
		t.Fatalf("ExportExcel: %v", err)
	}

	f, err := excelize.OpenFile(tmp.Name())
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("PotentialRequirements")
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	if len(rows) == 0 || len(rows[0]) < 14 {
		t.Fatalf("expected Enable column in header, got %v", rows)
	}
	if rows[0][13] != "Enable" {
		t.Fatalf("unexpected header %v", rows[0])
	}
	if len(rows) < 2 || len(rows[1]) < 14 || !strings.EqualFold(rows[1][13], "false") {
		t.Fatalf("expected default false in Enable column, got %v", rows[1])
	}
}

func TestImportPotentialRequirementsPromotion(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	header := []interface{}{"Field", "Value"}
	if err := f.SetSheetRow("Sheet1", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}
	if err := f.SetSheetRow("Sheet1", "A2", &[]interface{}{"Name", "Test"}); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}
	f.SetSheetName("Sheet1", "Project")

	f.NewSheet("Requirements")
	reqHeader := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags"}
	if err := f.SetSheetRow("Requirements", "A1", &reqHeader); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}

	f.NewSheet("PotentialRequirements")
	prHeader := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags", "Enable"}
	if err := f.SetSheetRow("PotentialRequirements", "A1", &prHeader); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}
	row := []interface{}{1, "Pot", "desc", 1, 1, "u", "Status", time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), 0, 0, "Cat", "tag", "true"}
	if err := f.SetSheetRow("PotentialRequirements", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}

	tmp, err := os.CreateTemp("", "imp-*.xlsx")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := f.SaveAs(tmp.Name()); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	pd, err := ImportProjectExcel(tmp.Name())
	if err != nil {
		t.Fatalf("ImportProjectExcel: %v", err)
	}
	if len(pd.Requirements) != 1 {
		t.Fatalf("expected requirement promoted, got %d", len(pd.Requirements))
	}
	if len(pd.PotentialRequirements) != 0 {
		t.Fatalf("expected no potential requirements, got %d", len(pd.PotentialRequirements))
	}
}
