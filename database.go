package PMFS

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pelletier/go-toml/v2"
	llm "github.com/rjboer/PMFS/pmfs/llm"
)

const (
	productsDir   = "products"
	indexFilename = "index.toml"
	projectTOML   = "project.toml"
	envBaseDir    = "PMFS_BASEDIR"
)

var (
	baseDir         = defaultBaseDir()
	baseProductsDir string
	indexPath       string

	ErrProductNotFound = errors.New("product not found")
	ErrProjectNotFound = errors.New("project not found")
)

func init() {
	setBaseDir(baseDir)
}

func defaultBaseDir() string {
	if dir := os.Getenv(envBaseDir); dir != "" {
		return dir
	}
	return "database"
}

// SetBaseDir overrides the base data directory and updates internal paths.
func SetBaseDir(dir string) {
	baseDir = dir
	setBaseDir(dir)
}

func setBaseDir(dir string) {
	baseProductsDir = filepath.Join(dir, productsDir)
	indexPath = filepath.Join(baseProductsDir, indexFilename)
}

// -----------------------------------------------------------------------------
// Memory model
// -----------------------------------------------------------------------------

// Database holds a list of products.
type Database struct {
	Products []Product `toml:"products"`
}

// Product represents a product with a set of projects.
type Product struct {
	ProductData
	Projects []Project `toml:"projects"`
}

// ProductData is the metadata stored for a product.
type ProductData struct {
	ID   int    `json:"id" toml:"id"`
	Name string `json:"name" toml:"name"`
}

// -----------------------------------------------------------------------------
// Filesystem helpers
// -----------------------------------------------------------------------------

// EnsureLayout creates base folder structure and ensures index.toml exists.
func EnsureLayout() error {
	if err := os.MkdirAll(baseProductsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", baseProductsDir, err)
	}
	if ok, err := fileExists(indexPath); err != nil {
		return err
	} else if !ok {
		idx := Database{Products: []Product{}}
		if err := writeTOML(indexPath, &idx); err != nil {
			return fmt.Errorf("write index.toml: %w", err)
		}
	}
	return nil
}

// Load reads index.toml into the shallow model.
func Load() (Database, error) {
	var idx Database
	if err := readTOML(indexPath, &idx); err != nil {
		if os.IsNotExist(err) {
			idx = Database{Products: []Product{}}
			if werr := writeTOML(indexPath, &idx); werr != nil {
				return idx, werr
			}
			return idx, nil
		}
		return idx, fmt.Errorf("read index.toml: %w", err)
	}
	if idx.Products == nil {
		idx.Products = []Product{}
	}
	return idx, nil
}

// Save writes the database index to disk.
func (db *Database) Save() error {
	if err := writeTOML(indexPath, db); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// NewProduct creates a product, writes index.toml, and returns its ID.
// ProductID = len(db.Products) + 1
func (db *Database) NewProduct(data ProductData) (int, error) {
	if data.Name == "" {
		return 0, errors.New("product name cannot be empty")
	}

	newID := len(db.Products) + 1
	pDir := productDir(newID)

	if err := os.MkdirAll(filepath.Join(pDir, "projects"), 0o755); err != nil {
		return 0, fmt.Errorf("mkdir product/projects: %w", err)
	}

	db.Products = append(db.Products, Product{
		ProductData: ProductData{ID: newID, Name: data.Name},
		Projects:    []Project{},
	})
	if err := db.Save(); err != nil {
		return 0, fmt.Errorf("error saving index, NewProduct: %w", err)
	}
	return newID, nil
}

// ModifyProduct updates existing product metadata and persists the index.
func (db *Database) ModifyProduct(data ProductData) (int, error) {
	for i := range db.Products {
		if db.Products[i].ID == data.ID {
			if data.Name != "" {
				db.Products[i].Name = data.Name
			}
			if err := db.Save(); err != nil {
				return 0, fmt.Errorf("error saving index, ModifyProduct: %w", err)
			}
			return data.ID, nil
		}
	}
	return 0, ErrProductNotFound
}

// CreateProject appends a project to the given product and writes its TOML.
// projectID = len(product.Projects) + 1
// db must be the database containing this product so the index can be persisted.
func (prd *Product) CreateProject(db *Database, projectName string) error {
	if projectName == "" {
		return errors.New("project name cannot be empty")
	}
	if db == nil {
		return errors.New("database cannot be nil")
	}

	newPrjID := len(prd.Projects) + 1
	prjDir := projectDir(prd.ID, newPrjID)
	if err := os.MkdirAll(prjDir, 0o755); err != nil {
		return fmt.Errorf("mkdir project dir: %w", err)
	}

	added := Project{ID: newPrjID, ProductID: prd.ID, Name: projectName, LLM: llm.DefaultClient}
	if err := added.Save(); err != nil {
		return fmt.Errorf("error saving TOML, CreateProject: %w", err)
	}

	prd.Projects = append(prd.Projects, added)
	if err := db.Save(); err != nil {
		return fmt.Errorf("error saving index, CreateProject: %w", err)
	}
	return nil
}

// LoadAllProjects loads all projects for all products in the database.
func (db *Database) LoadAllProjects() error {
	for i := range db.Products {
		if err := db.Products[i].LoadProjects(); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Small helpers
// -----------------------------------------------------------------------------

func productDir(productID int) string {
	return filepath.Join(baseProductsDir, strconv.Itoa(productID))
}

func projectDir(productID, projectID int) string {
	return filepath.Join(productDir(productID), "projects", strconv.Itoa(projectID))
}

func attachmentDir(productID, projectID int) string {
	return filepath.Join(productDir(productID), "projects", strconv.Itoa(projectID), "attachments")
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func writeTOML(path string, v any) error {
	data, err := toml.Marshal(v)
	if err != nil {
		return fmt.Errorf("toml marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func readTOML(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, v)
}
