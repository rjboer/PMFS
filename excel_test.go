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
	if len(rows) == 0 || len(rows[0]) < 18 {
		t.Fatalf("expected condition columns in header, got %v", rows)
	}
	if rows[0][13] != "Proposed" || rows[0][14] != "AIgenerated" || rows[0][15] != "AIanalyzed" || rows[0][16] != "Active" || rows[0][17] != "Deleted" {
		t.Fatalf("unexpected header %v", rows[0])
	}
	if len(rows) < 2 || len(rows[1]) < 18 || !strings.EqualFold(rows[1][13], "true") {
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
	reqHeader := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags", "Proposed", "AIgenerated", "AIanalyzed", "Active", "Deleted"}
	if err := f.SetSheetRow("Requirements", "A1", &reqHeader); err != nil {
		t.Fatalf("SetSheetRow: %v", err)
	}
	row := []interface{}{1, "Req", "desc", 1, 1, "u", "Status", time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), 0, 0, "Cat", "tag", "true", "false", "false", "false", "false"}
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

func TestExportImportDesignAspects(t *testing.T) {
	prj := &ProjectType{D: ProjectData{Requirements: []Requirement{{
		ID:            1,
		Name:          "Req",
		DesignAspects: []DesignAspect{{Name: "Aspect1", Description: "Desc1", Processed: true}},
	}}}}

	tmp, err := os.CreateTemp("", "da-*.xlsx")
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
	rows, err := f.GetRows("DesignAspects")
	f.Close()
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	if len(rows) < 2 || rows[1][1] != "Aspect1" {
		t.Fatalf("design aspects not exported: %v", rows)
	}

	pd, err := ImportProjectExcel(tmp.Name())
	if err != nil {
		t.Fatalf("ImportProjectExcel: %v", err)
	}
	if len(pd.Requirements) != 1 || len(pd.Requirements[0].DesignAspects) != 1 {
		t.Fatalf("design aspects not imported: %#v", pd.Requirements)
	}
	da := pd.Requirements[0].DesignAspects[0]
	if da.Name != "Aspect1" || !da.Processed {
		t.Fatalf("unexpected design aspect: %#v", da)
	}
}

func TestProjectImportExcelSyncByID(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	f := excelize.NewFile()
	defer f.Close()

	header := []interface{}{"Field", "Value"}
	_ = f.SetSheetRow("Sheet1", "A1", &header)
	_ = f.SetSheetRow("Sheet1", "A2", &[]interface{}{"Name", "Test"})
	f.SetSheetName("Sheet1", "Project")

	f.NewSheet("Requirements")
	reqHeader := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags", "Proposed", "AIgenerated", "AIanalyzed", "Active", "Deleted"}
	_ = f.SetSheetRow("Requirements", "A1", &reqHeader)
	row1 := []interface{}{1, "Updated", "d", 0, 0, "", "", now, now, 0, 0, "", "", "false", "false", "false", "false", "false"}
	row2 := []interface{}{0, "New", "d2", 0, 0, "", "", now, now, 0, 0, "", "", "false", "false", "false", "false", "false"}
	_ = f.SetSheetRow("Requirements", "A2", &row1)
	_ = f.SetSheetRow("Requirements", "A3", &row2)

	tmp, err := os.CreateTemp("", "merge-*.xlsx")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	if err := f.SaveAs(tmp.Name()); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	prj := &ProjectType{D: ProjectData{Requirements: []Requirement{{ID: 1, Name: "Old", CreatedAt: time.Now(), UpdatedAt: time.Now()}}}}
	if err := prj.ImportExcel(tmp.Name(), false); err != nil {
		t.Fatalf("ImportExcel: %v", err)
	}
	if len(prj.D.Requirements) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(prj.D.Requirements))
	}
	if prj.D.Requirements[0].Name != "Updated" {
		t.Fatalf("existing requirement not updated: %#v", prj.D.Requirements[0])
	}
	if prj.D.Requirements[1].ID != 2 || prj.D.Requirements[1].Name != "New" {
		t.Fatalf("new requirement not added: %#v", prj.D.Requirements[1])
	}
}

func TestProjectImportExcelReplace(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	f := excelize.NewFile()
	defer f.Close()

	header := []interface{}{"Field", "Value"}
	_ = f.SetSheetRow("Sheet1", "A1", &header)
	_ = f.SetSheetRow("Sheet1", "A2", &[]interface{}{"Name", "Test"})
	f.SetSheetName("Sheet1", "Project")

	f.NewSheet("Requirements")
	reqHeader := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags", "Proposed", "AIgenerated", "AIanalyzed", "Active", "Deleted"}
	_ = f.SetSheetRow("Requirements", "A1", &reqHeader)
	row := []interface{}{0, "New", "d", 0, 0, "", "", now, now, 0, 0, "", "", "false", "false", "false", "false", "false"}
	_ = f.SetSheetRow("Requirements", "A2", &row)

	tmp, err := os.CreateTemp("", "replace-*.xlsx")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	if err := f.SaveAs(tmp.Name()); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	prj := &ProjectType{D: ProjectData{Requirements: []Requirement{{ID: 1, Name: "Old"}}}}
	if err := prj.ImportExcel(tmp.Name(), true); err != nil {
		t.Fatalf("ImportExcel replace: %v", err)
	}
	if len(prj.D.Requirements) != 1 || prj.D.Requirements[0].Name != "New" {
		t.Fatalf("replace did not overwrite requirements: %#v", prj.D.Requirements)
	}
}
