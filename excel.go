package PMFS

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportExcel writes project data to an Excel workbook.
// It creates separate sheets for project metadata, requirements and other
// non-empty slices found on the project data. Attachments are currently
// skipped. The resulting workbook is saved to the provided path.
func (p *ProjectType) ExportExcel(path string) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	// Project metadata
	f.SetSheetName("Sheet1", "Project")
	header := []interface{}{"Field", "Value"}
	if err := f.SetSheetRow("Project", "A1", &header); err != nil {
		return err
	}
	rows := [][]interface{}{
		{"ID", p.ID},
		{"ProductID", p.ProductID},
		{"Name", p.Name},
		{"Scope", p.D.Scope},
		{"StartDate", p.D.StartDate.Format(time.RFC3339)},
		{"EndDate", p.D.EndDate.Format(time.RFC3339)},
		{"Status", p.D.Status},
		{"Priority", p.D.Priority},
	}
	for i, r := range rows {
		cell := fmt.Sprintf("A%d", i+2)
		if err := f.SetSheetRow("Project", cell, &r); err != nil {
			return err
		}
	}

	// Requirements sheet
	if len(p.D.Requirements) > 0 {
		sheet := "Requirements"
		f.NewSheet(sheet)
		header := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "Category", "Tags"}
		if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
			return err
		}
		for i, req := range p.D.Requirements {
			row := []interface{}{
				req.ID,
				req.Name,
				req.Description,
				req.Priority,
				req.Level,
				req.User,
				req.Status,
				req.CreatedAt.Format(time.RFC3339),
				req.UpdatedAt.Format(time.RFC3339),
				req.ParentID,
				req.Category,
				strings.Join(req.Tags, ","),
			}
			cell := fmt.Sprintf("A%d", i+2)
			if err := f.SetSheetRow(sheet, cell, &row); err != nil {
				return err
			}
		}
	}

	// Potential requirements sheet
	if len(p.D.PotentialRequirements) > 0 {
		sheet := "PotentialRequirements"
		f.NewSheet(sheet)
		header := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "Category", "Tags"}
		if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
			return err
		}
		for i, req := range p.D.PotentialRequirements {
			row := []interface{}{
				req.ID,
				req.Name,
				req.Description,
				req.Priority,
				req.Level,
				req.User,
				req.Status,
				req.CreatedAt.Format(time.RFC3339),
				req.UpdatedAt.Format(time.RFC3339),
				req.ParentID,
				req.Category,
				strings.Join(req.Tags, ","),
			}
			cell := fmt.Sprintf("A%d", i+2)
			if err := f.SetSheetRow(sheet, cell, &row); err != nil {
				return err
			}
		}
	}

	// Intelligence sheet
	if len(p.D.Intelligence) > 0 {
		sheet := "Intelligence"
		f.NewSheet(sheet)
		header := []interface{}{"ID", "Filepath", "Content", "Description", "ExtractedAt"}
		if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
			return err
		}
		for i, intel := range p.D.Intelligence {
			row := []interface{}{
				intel.ID,
				intel.Filepath,
				intel.Content,
				intel.Description,
				intel.ExtractedAt.Format(time.RFC3339),
			}
			cell := fmt.Sprintf("A%d", i+2)
			if err := f.SetSheetRow(sheet, cell, &row); err != nil {
				return err
			}
		}
	}

	// Save to path
	if err := f.SaveAs(path); err != nil {
		return err
	}
	return nil
}
