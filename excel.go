package PMFS

import (
	"fmt"
	"strconv"
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
		header := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags"}
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
				req.AttachmentIndex,
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
		header := []interface{}{"ID", "Name", "Description", "Priority", "Level", "User", "Status", "CreatedAt", "UpdatedAt", "ParentID", "AttachmentIndex", "Category", "Tags"}
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
				req.AttachmentIndex,
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

// ImportProjectExcel reads an Excel workbook and returns populated ProjectData.
// It expects a sheet named "Project" with key/value pairs for basic metadata
// and a "Requirements" sheet listing confirmed requirements. Additional
// optional sheets like "PotentialRequirements" and "Intelligence" are imported
// when present. Missing optional sheets are ignored.
func ImportProjectExcel(path string) (*ProjectData, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	var pd ProjectData

	// Project metadata
	rows, err := f.GetRows("Project")
	if err != nil {
		return nil, err
	}
	for _, r := range rows[1:] {
		if len(r) < 2 {
			continue
		}
		key, val := r[0], r[1]
		switch key {
		case "Name":
			pd.Name = val
		case "Scope":
			pd.Scope = val
		case "StartDate":
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return nil, err
			}
			pd.StartDate = t
		case "EndDate":
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return nil, err
			}
			pd.EndDate = t
		case "Status":
			pd.Status = val
		case "Priority":
			pd.Priority = val
		}
	}

	// Requirements
	reqRows, err := f.GetRows("Requirements")
	if err != nil {
		return nil, err
	}
	for _, row := range reqRows[1:] {
		if len(row) < 13 {
			continue
		}
		var req Requirement
		if req.ID, err = strconv.Atoi(row[0]); err != nil {
			return nil, err
		}
		req.Name = row[1]
		req.Description = row[2]
		if req.Priority, err = strconv.Atoi(row[3]); err != nil {
			return nil, err
		}
		if req.Level, err = strconv.Atoi(row[4]); err != nil {
			return nil, err
		}
		req.User = row[5]
		req.Status = row[6]
		if req.CreatedAt, err = time.Parse(time.RFC3339, row[7]); err != nil {
			return nil, err
		}
		if req.UpdatedAt, err = time.Parse(time.RFC3339, row[8]); err != nil {
			return nil, err
		}
		if req.ParentID, err = strconv.Atoi(row[9]); err != nil {
			return nil, err
		}
		if req.AttachmentIndex, err = strconv.Atoi(row[10]); err != nil {
			return nil, err
		}
		req.Category = row[11]
		if row[12] != "" {
			req.Tags = strings.Split(row[12], ",")
		}
		pd.Requirements = append(pd.Requirements, req)
	}

	// Potential requirements (optional)
	if prRows, err := f.GetRows("PotentialRequirements"); err == nil {
		for _, row := range prRows[1:] {
			if len(row) < 13 {
				continue
			}
			var req Requirement
			if req.ID, err = strconv.Atoi(row[0]); err != nil {
				return nil, err
			}
			req.Name = row[1]
			req.Description = row[2]
			if req.Priority, err = strconv.Atoi(row[3]); err != nil {
				return nil, err
			}
			if req.Level, err = strconv.Atoi(row[4]); err != nil {
				return nil, err
			}
			req.User = row[5]
			req.Status = row[6]
			if req.CreatedAt, err = time.Parse(time.RFC3339, row[7]); err != nil {
				return nil, err
			}
			if req.UpdatedAt, err = time.Parse(time.RFC3339, row[8]); err != nil {
				return nil, err
			}
			if req.ParentID, err = strconv.Atoi(row[9]); err != nil {
				return nil, err
			}
			if req.AttachmentIndex, err = strconv.Atoi(row[10]); err != nil {
				return nil, err
			}
			req.Category = row[11]
			if row[12] != "" {
				req.Tags = strings.Split(row[12], ",")
			}
			pd.PotentialRequirements = append(pd.PotentialRequirements, req)
		}
		pd.PotentialRequirements = Deduplicate(pd.PotentialRequirements)
	}

	// Intelligence (optional)
	if intelRows, err := f.GetRows("Intelligence"); err == nil {
		for _, row := range intelRows[1:] {
			if len(row) < 5 {
				continue
			}
			var intel Intelligence
			if intel.ID, err = strconv.Atoi(row[0]); err != nil {
				return nil, err
			}
			intel.Filepath = row[1]
			intel.Content = row[2]
			intel.Description = row[3]
			if intel.ExtractedAt, err = time.Parse(time.RFC3339, row[4]); err != nil {
				return nil, err
			}
			pd.Intelligence = append(pd.Intelligence, intel)
		}
	}

	return &pd, nil
}
