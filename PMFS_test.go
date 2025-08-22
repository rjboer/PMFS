package PMFS

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureLayoutUsesBaseDir(t *testing.T) {
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

func TestAddProductAndProject(t *testing.T) {
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
	prd := &idx.Products[0]
	if err := prd.AddProject(&idx, "prj1"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	prjToml := filepath.Join(dir, productsDir, "1", "projects", "1", projectTOML)
	if _, err := os.Stat(prjToml); err != nil {
		t.Fatalf("project toml missing: %v", err)
	}
}
