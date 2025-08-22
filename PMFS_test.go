package PMFS

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestEnsureLayoutCreatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	SetBaseDir(dir)
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	p := filepath.Join(dir, productsDir, indexFilename)
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("index not created at %s: %v", p, err)
	}
}

func TestAddProductCreatesDirAndUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	SetBaseDir(dir)
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	idx, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if err := idx.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	prodDir := filepath.Join(dir, productsDir, "1", "projects")
	if _, err := os.Stat(prodDir); err != nil {
		t.Fatalf("product dir missing: %v", err)
	}
	idx2, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if len(idx2.Products) != 1 || idx2.Products[0].Name != "prod1" {
		t.Fatalf("index not updated: %#v", idx2.Products)
	}
}

func TestAddProjectWritesTomlAndUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	SetBaseDir(dir)
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	idx, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if err := idx.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	// reload index to obtain product
	idx, err = LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	prd := &idx.Products[0]
	if err := prd.AddProject(&idx, "prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	prjToml := filepath.Join(dir, productsDir, "1", "projects", "1", projectTOML)
	if _, err := os.Stat(prjToml); err != nil {
		t.Fatalf("project toml missing: %v", err)
	}
	idx2, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if len(idx2.Products[0].Projects) != 1 || idx2.Products[0].Projects[0].Name != "prj1" {
		t.Fatalf("project not persisted to index: %#v", idx2.Products[0].Projects)
	}
}

func TestAddAttachmentFromInputMovesFileAndRecordsMetadata(t *testing.T) {
	// mock Gemini client to avoid external calls
	orig := gemini.SetClient(gemini.ClientFunc(func(path string) ([]gemini.Requirement, error) {
		return nil, nil
	}))
	defer gemini.SetClient(orig)

	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	SetBaseDir(dir)
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	idx, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if err := idx.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	idx, err = LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	prd := &idx.Products[0]
	if err := prd.AddProject(&idx, "prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	prj := &idx.Products[0].Projects[0]

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
	orig := gemini.SetClient(gemini.ClientFunc(func(path string) ([]gemini.Requirement, error) {
		return mockReqs, nil
	}))
	defer gemini.SetClient(orig)

	dir := t.TempDir()
	SetBaseDir(dir)
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	idx, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if err := idx.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	idx, err = LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	prd := &idx.Products[0]
	if err := prd.AddProject(&idx, "prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	prj := &idx.Products[0].Projects[0]

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

func TestIngestInputDirProcessesAllFiles(t *testing.T) {

	orig := gemini.SetClient(gemini.ClientFunc(func(path string) ([]gemini.Requirement, error) {
		return nil, nil
	}))
	defer gemini.SetClient(orig)

	t.Setenv("GEMINI_API_KEY", "test-key")

	dir := t.TempDir()
	SetBaseDir(dir)
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	idx, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if err := idx.AddProduct("prod1"); err != nil {
		t.Fatalf("AddProduct: %v", err)
	}
	idx, err = LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	prd := &idx.Products[0]
	if err := prd.AddProject(&idx, "prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	prj := &idx.Products[0].Projects[0]

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

	idxReload, err := LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if err := idxReload.LoadAllProjects(); err != nil {
		t.Fatalf("LoadAllProjects: %v", err)
	}
	if len(idxReload.Products) != 1 || len(idxReload.Products[0].Projects[0].D.Attachments) != len(files) {
		t.Fatalf("attachments not loaded via LoadAllProjects: %#v", idxReload.Products[0].Projects[0].D.Attachments)
	}
}
