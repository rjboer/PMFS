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

func TestNewProductCreatesDirAndUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}

	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)

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
	if len(db2.Products) != 1 || db2.Products[0].Name != "prod1" {
		t.Fatalf("index not updated: %#v", db2.Products)
	}
}

func TestModifyProductUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)

	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	id, err := db.NewProduct(ProductData{Name: "prod1"})
	if err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	if _, err := db.ModifyProduct(ProductData{ID: id, Name: "prod1-upd"}); err != nil {
		t.Fatalf("ModifyProduct: %v", err)
	}
	db2, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if db2.Products[0].Name != "prod1-upd" {
		t.Fatalf("product not updated: %#v", db2.Products[0])
	}
}

func TestNewProjectWritesTomlAndUpdatesIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]
	if _, err := prd.NewProject(ProjectData{Name: "prj1"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)

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
	if len(db2.Products[0].Projects) != 1 || db2.Products[0].Projects[0].Name != "prj1" {
		t.Fatalf("project not persisted to index: %#v", db2.Products[0].Projects)
	}
}

func TestModifyProjectUpdatesTomlAndIndex(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}

	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]
	id, err := prd.NewProject(ProjectData{Name: "prj1"})
	if err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if _, err := prd.ModifyProject(id, ProjectData{Name: "prj1-upd"}); err != nil {
		t.Fatalf("ModifyProject: %v", err)
	}

	db2, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if db2.Products[0].Projects[0].Name != "prj1-upd" {
		t.Fatalf("project not updated in index: %#v", db2.Products[0].Projects[0])
	}

	prjReload := ProjectType{ID: id, ProductID: prd.ID}
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if prjReload.Name != "prj1-upd" {
		t.Fatalf("project toml not updated: %s", prjReload.Name)
	}

	if db2.LLM == nil {

		t.Fatalf("LLM not set to default")
	}
}

func TestAddAttachmentFromInputMovesFileAndRecordsMetadata(t *testing.T) {
	// mock Gemini client to avoid external calls
	askCalls := 0
	orig := llm.SetClient(gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return nil, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if askCalls%2 == 0 {
				askCalls++
				return "summary", nil
			}
			askCalls++
			return "[]", nil
		},
	})
	defer llm.SetClient(orig)

	t.Setenv("GEMINI_API_KEY", "test-key")
	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}

	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}

	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]

	if _, err := prd.NewProject(ProjectData{Name: "prj1"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prj := &db.Products[0].Projects[0]

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
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Attachments) != 1 || prjReload.D.Attachments[0].Filename != fname {
		t.Fatalf("attachment not persisted: %#v", prjReload.D.Attachments)
	}
	if !prjReload.D.Attachments[0].Analyzed {
		t.Fatalf("Analyzed flag not persisted: %#v", prjReload.D.Attachments[0])
	}
}

func TestAttachmentGenerateRequirements(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	mockReqs := []gemini.Requirement{{Name: "R1", Description: "D1"}}
	dir := t.TempDir()

	var expected string
	askCalls := 0
	orig := llm.SetClient(gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			if path != expected {
				t.Fatalf("AnalyzeAttachment called with %s, want %s", path, expected)
			}
			return mockReqs, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if askCalls == 0 {
				askCalls++
				return "summary", nil
			}
			askCalls++
			return `[{"name":"Aspect1","description":"Desc1"}]`, nil
		},
	})
	defer llm.SetClient(orig)

	if _, err := LoadSetup(dir); err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}

	prj := &ProjectType{ProductID: 1, ID: 2}
	att := Attachment{RelPath: filepath.ToSlash(filepath.Join("attachments", "1", "f.txt"))}
	prj.D.Attachments = []Attachment{att}
	ptr := &prj.D.Attachments[0]
	expected = filepath.Join(projectDir(prj.ProductID, prj.ID), att.RelPath)
	if err := os.MkdirAll(filepath.Dir(expected), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(expected, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := ptr.GenerateRequirements(prj, ""); err != nil {
		t.Fatalf("GenerateRequirements: %v", err)
	}
	if !ptr.Analyzed {
		t.Fatalf("attachment not marked analyzed")
	}
	if len(prj.D.Requirements) != 1 || prj.D.Requirements[0].Name != "R1" {
		t.Fatalf("unexpected requirements: %#v", prj.D.Requirements)
	}
	if prj.D.Requirements[0].AttachmentIndex != 0 {
		t.Fatalf("attachment index not set: %#v", prj.D.Requirements[0])
	}
	if !prj.D.Requirements[0].Condition.Proposed || !prj.D.Requirements[0].Condition.AIgenerated {
		t.Fatalf("condition flags not set: %#v", prj.D.Requirements[0].Condition)
	}
	if len(prj.D.Intelligence) != 1 || prj.D.Intelligence[0].Description != "summary" {
		t.Fatalf("intelligence not generated: %#v", prj.D.Intelligence)
	}
	if len(prj.D.Intelligence[0].DesignAngles) != 1 || prj.D.Intelligence[0].DesignAngles[0].Name != "Aspect1" {
		t.Fatalf("design aspects missing: %#v", prj.D.Intelligence[0].DesignAngles)
	}

	var dp struct {
		D ProjectData `toml:"projectdata"`
	}
	p := filepath.Join(projectDir(prj.ProductID, prj.ID), projectTOML)
	if err := readTOML(p, &dp); err != nil {
		t.Fatalf("readTOML: %v", err)
	}
	if len(dp.D.Requirements) != 1 {
		t.Fatalf("project.toml not updated: %#v", dp.D.Requirements)
	}
	if dp.D.Requirements[0].AttachmentIndex != 0 {
		t.Fatalf("attachment index not persisted: %#v", dp.D.Requirements[0])
	}
	if len(dp.D.Intelligence) != 1 {
		t.Fatalf("intelligence not persisted: %#v", dp.D.Intelligence)
	}
	if askCalls != 2 {
		t.Fatalf("unexpected number of Ask calls: %d", askCalls)
	}
}

func TestAddAttachmentAnalyzesAndAppendsRequirements(t *testing.T) {
	mockReqs := []gemini.Requirement{{ID: 1, Name: "R1", Description: "D1"}}
	askCalls := 0
	orig := llm.SetClient(gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return mockReqs, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if askCalls%2 == 0 {
				askCalls++
				return "summary", nil
			}
			askCalls++
			return "[]", nil
		},
	})
	defer llm.SetClient(orig)

	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}

	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}

	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]

	if _, err := prd.NewProject(ProjectData{Name: "prj1"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	prj := &db.Products[0].Projects[0]

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
	if len(prj.D.Requirements) != len(mockReqs) {
		t.Fatalf("expected %d requirements, got %d", len(mockReqs), len(prj.D.Requirements))
	}
	if prj.D.Requirements[0].Name != mockReqs[0].Name {
		t.Fatalf("requirements not appended")
	}
	if prj.D.Requirements[0].AttachmentIndex != 0 {
		t.Fatalf("attachment index not set: %#v", prj.D.Requirements[0])
	}
	// ensure requirements persisted to disk
	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Requirements) != len(mockReqs) {
		t.Fatalf("expected %d persisted requirements, got %d", len(mockReqs), len(prjReload.D.Requirements))
	}
	if prjReload.D.Requirements[0].Name != mockReqs[0].Name {
		t.Fatalf("requirements not persisted: %#v", prjReload.D.Requirements)
	}
	if prjReload.D.Requirements[0].AttachmentIndex != 0 {
		t.Fatalf("attachment index not persisted: %#v", prjReload.D.Requirements[0])
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

	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]
	if _, err := prd.NewProject(ProjectData{Name: "prj1"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	prj := &db.Products[0].Projects[0]

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
	if len(prj.D.Requirements) == 0 {
		t.Fatalf("no requirements returned")
	}

	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Requirements) == 0 {
		t.Fatalf("requirements not persisted: %#v", prjReload.D.Requirements)
	}
	if prjReload.D.Requirements[0].Name == "" {
		t.Fatalf("empty requirement name: %#v", prjReload.D.Requirements[0])
	}
}

func TestIngestInputDirProcessesAllFiles(t *testing.T) {
	askCalls := 0
	orig := llm.SetClient(gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return nil, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if askCalls%2 == 0 {
				askCalls++
				return "summary", nil
			}
			askCalls++
			return "[]", nil
		},
	})
	defer llm.SetClient(orig)

	t.Setenv("GEMINI_API_KEY", "test-key")

	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}

	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}
	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]
	if _, err := prd.NewProject(ProjectData{Name: "prj1"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	prj := &db.Products[0].Projects[0]

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
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Attachments) != len(files) {
		t.Fatalf("expected %d attachments persisted, got %d", len(files), len(prjReload.D.Attachments))
	}

	dbReload, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	if err := dbReload.LoadAllProjects(); err != nil {
		t.Fatalf("LoadAllProjects: %v", err)
	}
	if len(dbReload.Products) != 1 || len(dbReload.Products[0].Projects[0].D.Attachments) != len(files) {
		t.Fatalf("attachments not loaded via LoadAllProjects: %#v", dbReload.Products[0].Projects[0].D.Attachments)
	}
}

func TestAttachmentManagerAddFromInputFolder(t *testing.T) {
	askCalls := 0
	orig := llm.SetClient(gemini.ClientFunc{
		AnalyzeAttachmentFunc: func(path string) ([]gemini.Requirement, error) {
			return nil, nil
		},
		AskFunc: func(prompt string) (string, error) {
			if askCalls%2 == 0 {
				askCalls++
				return "summary", nil
			}
			askCalls++
			return "[]", nil
		},
	})
	defer llm.SetClient(orig)

	dir := t.TempDir()
	db, err := LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}

	if _, err := db.NewProduct(ProductData{Name: "prod1"}); err != nil {
		t.Fatalf("NewProduct: %v", err)
	}

	db, err = LoadSetup(dir)
	if err != nil {
		t.Fatalf("LoadSetup: %v", err)
	}
	prd := &db.Products[0]
	if _, err := prd.NewProject(ProjectData{Name: "prj1"}); err != nil {
		t.Fatalf("NewProject: %v", err)
	}
	if err := db.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	prj := &db.Products[0].Projects[0]

	prjDir := filepath.Join(dir, productsDir, "1", "projects", "1")
	inputDir := filepath.Join(prjDir, "input")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("Mkdir input: %v", err)
	}
	files := []string{"a.txt", "b.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(inputDir, f), []byte(f), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", f, err)
		}
	}

	atts, err := prj.Attachments().AddFromInputFolder()
	if err != nil {
		t.Fatalf("AddFromInputFolder: %v", err)
	}
	if len(atts) != len(files) {
		t.Fatalf("expected %d attachments, got %d", len(files), len(atts))
	}

	for i, name := range files {
		dst := filepath.Join(prjDir, "attachments", strconv.Itoa(i+1), name)
		if _, err := os.Stat(dst); err != nil {
			t.Fatalf("missing moved file %s: %v", dst, err)
		}
		if _, err := os.Stat(filepath.Join(inputDir, name)); !os.IsNotExist(err) {
			t.Fatalf("source file %s still exists", name)
		}
	}

	prjReload := ProjectType{ID: prj.ID, ProductID: prj.ProductID}
	if err := prjReload.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(prjReload.D.Attachments) != len(files) {
		t.Fatalf("expected %d attachments persisted, got %d", len(files), len(prjReload.D.Attachments))
	}
}
