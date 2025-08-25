package PMFS

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	llm "github.com/rjboer/PMFS/pmfs/llm"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestLoadSetupCreatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	p := filepath.Join(dir, productsDir, indexFilename)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("index not created at %s: %v", p, err)
	}
}

func TestAddProductCreatesDirAndUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prodDir := filepath.Join(dir, productsDir, "1", "projects")
	if _, err := os.Stat(prodDir); err != nil {
		t.Fatalf("product dir missing: %v", err)
	}
	db2, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if len(db2.Index.Products) != 1 || db2.Index.Products[0].Name != "prod1" {
		t.Fatalf("index not updated: %#v", db2.Index.Products)
	}
}

func TestAddProjectWritesTomlAndUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Index.Products[0]
	if _, err := prd.AddProject("prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prjToml := filepath.Join(dir, productsDir, "1", "projects", "1", projectTOML)
	if _, err := os.Stat(prjToml); err != nil {
		t.Fatalf("project toml missing: %v", err)
	}
	db2, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if len(db2.Index.Products[0].Projects) != 1 || db2.Index.Products[0].Projects[0].Name != "prj1" {
		t.Fatalf("project not persisted to index: %#v", db2.Index.Products[0].Projects)
	}
}

func TestAddAttachmentFromInputMovesFileAndRecordsMetadata(t *testing.T) {
	// mock Gemini client to avoid external calls
	orig := llm.SetClient(gemini.ClientFunc{AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
		return nil, nil
	}})
	defer llm.SetClient(orig)

	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Index.Products[0]
	if _, err := prd.AddProject("prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prj := &db.Index.Products[0].Projects[0]

	inputDir := filepath.Join(dir, "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("Mkdir input: %v", err)
	}
	fname := "sample.txt"
	src := filepath.Join(inputDir, fname)
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	att, err := prj.AddAttachmentFromInput(inputDir, fname)
	if err != nil {
		t.Fatalf("AddAttachmentFromInput: %v", err)
	}

	dst := filepath.Join(dir, productsDir, "1", "projects", "1", "attachments", "1", fname)
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("attachment not moved: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source file still exists")
	}
	if len(prj.D.Attachments) != 1 || prj.D.Attachments[0] != att {
		t.Fatalf("attachment metadata not recorded: %#v", prj.D.Attachments)
	}
	if !att.Analyzed {
		t.Fatalf("attachment not analyzed")
	}
	// ensure metadata persisted to project.toml
	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.LoadProject(); err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if len(prjReload.D.Attachments) != 1 || prjReload.D.Attachments[0].Filename != fname {
		t.Fatalf("attachment not persisted: %#v", prjReload.D.Attachments)
	}
	if !prjReload.D.Attachments[0].Analyzed {
		t.Fatalf("Analyzed flag not persisted: %#v", prjReload.D.Attachments[0])
	}
}

func TestAddAttachmentAnalyzesAndAppendsRequirements(t *testing.T) {
	mockReqs := []gemini.Requirement{{ID: 1, Name: "R1", Description: "D1"}}
	orig := llm.SetClient(gemini.ClientFunc{AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
		return mockReqs, nil
	}})
	defer llm.SetClient(orig)

	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Index.Products[0]
	if _, err := prd.AddProject("prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prj := &db.Index.Products[0].Projects[0]

	inputDir := filepath.Join(dir, "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("Mkdir input: %v", err)
	}
	fname := "sample.txt"
	src := filepath.Join(inputDir, fname)
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	att, err := prj.AddAttachmentFromInput(inputDir, fname)
	if err != nil {
		t.Fatalf("AddAttachmentFromInput: %v", err)
	}
	// file moved to project attachments directory
	dst := filepath.Join(dir, productsDir, "1", "projects", "1", "attachments", "1", fname)
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source file still exists")
	}
	// attachment metadata recorded in project
	if len(prj.D.Attachments) != 1 || prj.D.Attachments[0] != att {
		t.Fatalf("attachment metadata not recorded: %#v", prj.D.Attachments)
	}
	if !att.Analyzed {
		t.Fatalf("attachment not analyzed")
	}
	if len(prj.D.PotentialRequirements) != len(mockReqs) {
		t.Fatalf("expected %d potential requirements, got %d", len(mockReqs), len(prj.D.PotentialRequirements))
	}
	if prj.D.PotentialRequirements[0].Name != mockReqs[0].Name {
		t.Fatalf("requirements not appended")
	}
	// ensure requirements persisted to disk
	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.LoadProject(); err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if len(prjReload.D.PotentialRequirements) != len(mockReqs) {
		t.Fatalf("expected %d persisted requirements, got %d", len(mockReqs), len(prjReload.D.PotentialRequirements))
	}
	if prjReload.D.PotentialRequirements[0].Name != mockReqs[0].Name {
		t.Fatalf("requirements not persisted: %#v", prjReload.D.PotentialRequirements)
	}
}

func TestAddAttachmentRealAPI(t *testing.T) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		t.Skip("GEMINI_API_KEY not set")
	}
	t.Setenv("GEMINI_API_KEY", key)

	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Index.Products[0]
	if _, err := prd.AddProject("prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prj := &db.Index.Products[0].Projects[0]

	inputDir := filepath.Join(dir, "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("Mkdir input: %v", err)
	}
	fname := "req.txt"
	src := filepath.Join(inputDir, fname)
	content := "The system shall allow users to upload files."
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	att, err := prj.AddAttachmentFromInput(inputDir, fname)
	if err != nil {
		t.Fatalf("AddAttachmentFromInput: %v", err)
	}

	dst := filepath.Join(dir, productsDir, "1", "projects", "1", "attachments", "1", fname)
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("attachment not moved: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source file still exists")
	}
	if len(prj.D.Attachments) != 1 || prj.D.Attachments[0] != att {
		t.Fatalf("attachment metadata not recorded: %#v", prj.D.Attachments)
	}
	if !att.Analyzed {
		t.Fatalf("attachment not analyzed")
	}
	if len(prj.D.PotentialRequirements) == 0 {
		t.Fatalf("no requirements returned")
	}

	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.LoadProject(); err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if len(prjReload.D.PotentialRequirements) == 0 {
		t.Fatalf("requirements not persisted: %#v", prjReload.D.PotentialRequirements)
	}
	if prjReload.D.PotentialRequirements[0].Name == "" {
		t.Fatalf("empty requirement name: %#v", prjReload.D.PotentialRequirements[0])
	}
}

func TestIngestInputDirProcessesAllFiles(t *testing.T) {

	orig := llm.SetClient(gemini.ClientFunc{AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
		return nil, nil
	}})
	defer llm.SetClient(orig)

	t.Setenv("GEMINI_API_KEY", "test-key")

	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Index.Products[0]
	if _, err := prd.AddProject("prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prj := &db.Index.Products[0].Projects[0]

	inputDir := filepath.Join(dir, "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("Mkdir input: %v", err)
	}
	files := []string{"a.txt", "b.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(inputDir, f), []byte(f), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", f, err)
		}
	}

	atts, err := prj.IngestInputDir(inputDir)
	if err != nil {
		t.Fatalf("IngestInputDir: %v", err)
	}
	if len(atts) != len(files) {
		t.Fatalf("expected %d attachments, got %d", len(files), len(atts))
	}

	for i, name := range files {
		dst := filepath.Join(dir, productsDir, "1", "projects", "1", "attachments", strconv.Itoa(i+1), name)
		if _, err := os.Stat(dst); err != nil {
			t.Fatalf("missing moved file %s: %v", dst, err)
		}
		if _, err := os.Stat(filepath.Join(inputDir, name)); !os.IsNotExist(err) {
			t.Fatalf("source file %s still exists", name)
		}
	}

	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.LoadProject(); err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if len(prjReload.D.Attachments) != len(files) {
		t.Fatalf("expected %d attachments persisted, got %d", len(files), len(prjReload.D.Attachments))
	}

	dbReload, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if err := dbReload.Index.LoadAllProjects(); err != nil {
		t.Fatalf("LoadAllProjects: %v", err)
	}
	if len(dbReload.Index.Products) != 1 || len(dbReload.Index.Products[0].Projects[0].D.Attachments) != len(files) {
		t.Fatalf("attachments not loaded via LoadAllProjects: %#v", dbReload.Index.Products[0].Projects[0].D.Attachments)
	}
}
