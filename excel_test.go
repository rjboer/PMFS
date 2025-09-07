package PMFS

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestExportRequirementsConditionColumns(t *testing.T) {
	prj := &ProjectType{D: ProjectData{Requirements: []Requirement{{
		ID:        1,
		Name:      "Req",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Condition: ConditionType{Proposed: true},
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

	rows, err := f.GetRows("Requirements")
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	if len(rows) == 0 || len(rows[0]) < 17 {
		t.Fatalf("expected condition columns in header, got %v", rows)
	}
	if rows[0][13] != "Proposed" || rows[0][14] != "AIgenerated" || rows[0][15] != "Active" || rows[0][16] != "Deleted" {
		t.Fatalf("unexpected header %v", rows[0])
	}
	if len(rows) < 2 || len(rows[1]) < 17 || !strings.EqualFold(rows[1][13], "true") {
		t.Fatalf("expected proposed true in row, got %v", rows[1])
	}
}

func TestImportRequirements(t *testing.T) {
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
	reqHeader := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags", "Proposed", "AIgenerated", "Active", "Deleted"}
	if err := f.SetSheetRow("Requirements", "A1", &reqHeader); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}
	row := []interface{}{1, "Req", "desc", 1, 1, "u", "Status", time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), 0, 0, "Cat", "tag", "true", "false", "false", "false"}
	if err := f.SetSheetRow("Requirements", "A2", &row); err != nil {
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
		t.Fatalf("expected one requirement, got %d", len(pd.Requirements))
	}
	if !pd.Requirements[0].Condition.Proposed {
		t.Fatalf("condition not imported: %#v", pd.Requirements[0].Condition)
	}
}
