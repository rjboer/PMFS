package PMFS

import (
	"os"
	"path/filepath"
	"sort"
)

// AttachmentManager provides helper methods for managing attachments of a project.
type AttachmentManager struct {
	prj *ProjectType
	db  *Database
}

// Attachments returns an AttachmentManager for this project.
func (prj *ProjectType) Attachments(db *Database) AttachmentManager {
	return AttachmentManager{prj: prj, db: db}
}

// AddFromInputFolder scans the project's default "input" directory and ingests
// all regular files into the project's attachments directory.
// Files are moved into attachments/ and project.toml is updated.
func (am AttachmentManager) AddFromInputFolder() ([]Attachment, error) {
	inputDir := filepath.Join(projectDir(am.prj.ProductID, am.prj.ID), "input")

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) == 0 || name[0] == '.' {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	ingested := make([]Attachment, 0, len(names))
	for _, n := range names {
		att, err := am.prj.AddAttachmentFromInput(am.db, inputDir, n)
		if err != nil {
			return ingested, err
		}
		ingested = append(ingested, att)
	}
	return ingested, nil
}
