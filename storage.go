package storage

import (
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// -----------------------------------------------------------------------------
// Constants & paths
// -----------------------------------------------------------------------------

const (
	databaseDir   = "database"
	productsDir   = "products"
	indexFilename = "index.toml"
	projectTOML   = "project.toml"
)

var (
	baseProductsDir = filepath.Join(databaseDir, productsDir)
	indexPath       = filepath.Join(baseProductsDir, indexFilename)

	ErrProductNotFound = errors.New("product not found")
	ErrProjectNotFound = errors.New("project not found")
)

// -----------------------------------------------------------------------------
// Memory model
// -----------------------------------------------------------------------------

// Index is a minimal placeholder for products list.
// Next IDs are derived from len(products)+1.
type Index struct {
	Products []ProductType `toml:"products"`
}

type ProductType struct {
	ID       int           `toml:"id"`
	Name     string        `toml:"name"`
	Projects []ProjectType `toml:"projects"`
}

// ProjectFile is the project's memory model placeholder.
type ProjectType struct {
	ID        int         `json:"id" toml:"id"`
	ProductID int         `json:"productid" toml:"productid"`
	Name      string      `json:"name" toml:"name"`
	D         ProjectData `json:"projectdata" toml:"projectdata"`
}

type ProjectData struct {
	Scope        string        `json:"scope" toml:"scope"`
	StartDate    time.Time     `json:"start_date" toml:"start_date"`
	EndDate      time.Time     `json:"end_date" toml:"end_date"`
	Status       string        `json:"status" toml:"status"`
	Priority     string        `json:"priority" toml:"priority"`
	Requirements []Requirement `json:"requirements" toml:"requirements"`
	//	Attachments  []Attachment       `json:"attachments" toml:"attachments"`
	// Intelligence holds data extracted from attachments (e.g., screenshots, documents).
	//	Intelligence []Intelligence `json:"intelligence" toml:"intelligence"`
	// IntelligenceLinks connect extracted intelligence with confirmed requirements.
	//	IntelligenceLinks []IntelligenceLink `json:"intelligence_links" toml:"intelligence_links"`
	// PotentialRequirements are proposed requirements derived from intelligence analysis.
	//	PotentialRequirements []PotentialRequirement `json:"potential_requirements" toml:"potential_requirements"`
	// RequirementRelations holds the LLM-scored relationships between requirements.
	//	RequirementRelations  []RequirementRelation `json:"requirement_relations" toml:"requirement_relations"`
	//	RequirementCategories []string              `json:"requirement_Categories" toml:"requirement_Categories"`
	FixedCategories bool `json:"requirement_FixedCategories" toml:"requirement_FixedCategories"`
}

// Requirement represents a confirmed requirement with detailed metadata.
type Requirement struct {
	ID          int       `json:"id" toml:"id"`
	Name        string    `json:"name" toml:"name"`
	Description string    `json:"description" toml:"description"`
	Priority    int       `json:"priority" toml:"priority"` // 1 (highest) to 8 (lowest)
	Level       int       `json:"level" toml:"level"`       // Hierarchical level within requirements.
	Owner       string    `json:"owner" toml:"owner"`       // Consider replacing with a User struct for richer user data.
	Status      string    `json:"status" toml:"status"`     // e.g., "Draft", "Confirmed"
	CreatedAt   time.Time `json:"created_at" toml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" toml:"updated_at"`
	ParentID    int       `json:"parent_id" toml:"parent_id"` // 0 for top‑level
	Category    string    `json:"category" toml:"category"`   // e.g., "System Requirements"
	//History     []ChangeLog `json:"history" toml:"history"`     // Record of changes to the requirement.
	// Optional: Tags can help with flexible categorization or filtering.
	//Tags []string `json:"tags,omitempty" toml:"tags"`
}

// -----------------------------------------------------------------------------
// Init helpers
// -----------------------------------------------------------------------------
//
// EnsureLayout creates base folder structure and ensures index.toml exists.
func EnsureLayout() error {
	if err := os.MkdirAll(baseProductsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", baseProductsDir, err)
	}
	// Create empty index if missing.
	if ok, err := fileExists(indexPath); err != nil {
		return err
	} else if !ok {
		idx := Index{Products: []ProductType{}}
		if err := writeTOML(indexPath, &idx); err != nil {
			return fmt.Errorf("write index.toml: %w", err)
		}
	}
	return nil
}

// LoadIndex reads index.toml into the shallow model.
func LoadIndex() (Index, error) {
	var idx Index
	if err := readTOML(indexPath, &idx); err != nil {
		if os.IsNotExist(err) {
			// Create a fresh one if missing (keeps flow simple)
			idx = Index{Products: []ProductType{}}
			if werr := writeTOML(indexPath, &idx); werr != nil {
				return idx, werr
			}
			return idx, nil
		}
		return idx, fmt.Errorf("read index.toml: %w", err)
	}
	if idx.Products == nil {
		idx.Products = []ProductType{}
	}
	return idx, nil
}

// -----------------------------------------------------------------------------
// Public ops (global/exported)
// -----------------------------------------------------------------------------

// AddProduct appends a product to the index and creates its directory skeleton.
// ProductID = len(idx.Products) + 1
func (idx *Index) AddProduct(name string) error {
	if name == "" {
		return errors.New("product name cannot be empty")
	}

	newID := len(idx.Products) + 1
	pDir := productDir(newID)

	// Create product/<id>/projects (idempotent)
	if err := os.MkdirAll(filepath.Join(pDir, "projects"), 0o755); err != nil {
		return fmt.Errorf("mkdir product/projects: %w", err)
	}

	// Update in-memory index (shallow placeholder)
	idx.Products = append(idx.Products, ProductType{
		ID:       newID,
		Name:     name,
		Projects: []ProjectType{},
	})
	if err := idx.SaveIndex(); err != nil {
		return fmt.Errorf("Error Save-ing Toml, AddProduct function: %w", err)
	}

	return nil
}

func (idx *Index) SaveIndex() error {
	// Persist index
	idx2 := *idx
	for i, _ := range idx2.Products {
		for i2, _ := range idx2.Products[i].Projects {
			//empty out projectdata... otherwise it will be too big.
			//project specifics are stored in a project
			idx2.Products[i].Projects[i2].D = ProjectData{}
		}
	}
	if err := writeTOML(indexPath, &idx2); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// AddProject appends a project to the given product and writes its TOML.
// projectID = len(product.Projects) + 1
// Collisions on disk are acceptable by your policy (we overwrite TOML).
//
// idx must be the index containing this product so the index can be
// persisted after adding the project.
func (prd *ProductType) AddProject(idx *Index, projectName string) error {
	if projectName == "" {
		return errors.New("project name cannot be empty")
	}
	if idx == nil {
		return errors.New("index cannot be nil")
	}

	newPrjID := len(prd.Projects) + 1
	prjDir := projectDir(prd.ID, newPrjID)

	// Ensure dir (idempotent)
	if err := os.MkdirAll(prjDir, 0o755); err != nil {
		return fmt.Errorf("mkdir project dir: %w", err)
	}

	addedproject := ProjectType{
		ID:        newPrjID,
		ProductID: prd.ID,
		Name:      projectName,
	}

	if err := addedproject.SaveProject(); err != nil {
		return fmt.Errorf("Error Save-ing Toml, AddProject function: %w", err)
	}

	// Update in-memory index and persist
	prd.Projects = append(prd.Projects, addedproject)
	if err := idx.SaveIndex(); err != nil {
		return fmt.Errorf("Error Save-ing Index, AddProject function: %w", err)
	}

	return nil
}

func (prj *ProjectType) SaveProject() error {
	// Persist index
	prjDir := projectDir(prj.ProductID, prj.ID)
	tomlPath := filepath.Join(prjDir, projectTOML)

	if err := writeTOML(tomlPath, prj); err != nil {
		return fmt.Errorf("error write-ing project toml: %w", err)
	}
	return nil
}

// LoadProject loads a single project's TOML for this product.
func (prj *ProjectType) LoadProject() error {

	prjDir := projectDir(prj.ProductID, prj.ID)
	tomlPath := filepath.Join(prjDir, projectTOML)

	if err := readTOML(tomlPath, prj); err != nil {
		if os.IsNotExist(err) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("Error while reading a project, read %s: %w", tomlPath, err)
	}
	return nil
}

// LoadProjects loads all listed projects for this product by reading each project.toml.
func (prd *ProductType) LoadProjects() error {
	if prd.Projects == nil || len(prd.Projects) == 0 {
		return nil
	}
	for i := range prd.Projects {
		prd.Projects[i].ProductID = prd.ID
		err := prd.Projects[i].LoadProject()
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadAllProjects loads all projects for all products in the index.
func (idx *Index) LoadAllProjects() error {
	for i := range idx.Products {
		err := idx.Products[i].LoadProjects()
		if err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Small helpers
// because it is quite easy to do wrong!
//
// -----------------------------------------------------------------------------

func productDir(productID int) string {
	return filepath.Join(baseProductsDir, strconv.Itoa(productID))
}

func projectDir(productID, projectID int) string {
	return filepath.Join(productDir(productID), "projects", strconv.Itoa(projectID))
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

// (Optional) helper if you ever need numeric subdirs; kept simple & unused here.
func numericSubdirs(dir string) ([]int, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, err
	}
	out := make([]int, 0, len(ents))
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		if id, err := strconv.Atoi(e.Name()); err == nil && id > 0 {
			out = append(out, id)
		}
	}
	sort.Ints(out)
	return out, nil
}
