package PMFS

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	_ "github.com/rjboer/PMFS/internal/config"
	gates "github.com/rjboer/PMFS/pmfs/llm/gates"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/interact"
)

// -----------------------------------------------------------------------------
// Constants & paths
// -----------------------------------------------------------------------------

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

// ProjectType is the project's memory model placeholder.
type ProjectType struct {
	ID        int    `json:"id" toml:"id"`
	ProductID int    `json:"productid" toml:"productid"`
	Name      string `json:"name" toml:"name"`
	// D contains the heavy project data and is stored only in each
	// project's individual TOML file. The field is skipped when the
	// index is written to disk so the index remains lightweight.
	D ProjectData `json:"projectdata" toml:"-"`
}

type ProjectData struct {
	Scope        string        `json:"scope" toml:"scope"`
	StartDate    time.Time     `json:"start_date" toml:"start_date"`
	EndDate      time.Time     `json:"end_date" toml:"end_date"`
	Status       string        `json:"status" toml:"status"`
	Priority     string        `json:"priority" toml:"priority"`
	Requirements []Requirement `json:"requirements" toml:"requirements"`
	Attachments  []Attachment  `json:"attachments" toml:"attachments"`
	// Intelligence holds data extracted from attachments (e.g., screenshots, documents).
	Intelligence []Intelligence `json:"intelligence" toml:"intelligence"`
	// IntelligenceLinks connect extracted intelligence with confirmed requirements.
	//	IntelligenceLinks []IntelligenceLink `json:"intelligence_links" toml:"intelligence_links"`
	// PotentialRequirements are proposed requirements derived from intelligence analysis.
	PotentialRequirements []Requirement `json:"potential_requirements" toml:"potential_requirements"`
	// RequirementRelations holds the LLM-scored relationships between requirements.
	//	RequirementRelations  []RequirementRelation `json:"requirement_relations" toml:"requirement_relations"`
	//	RequirementCategories []string              `json:"requirement_Categories" toml:"requirement_Categories"`
	FixedCategories bool `json:"requirement_FixedCategories" toml:"requirement_FixedCategories"`
}

// Requirement represents a confirmed requirement with detailed metadata.
type Requirement struct {
	ID               int             `json:"id" toml:"id"`
	Name             string          `json:"name" toml:"name"`
	Description      string          `json:"description" toml:"description"`
	Priority         int             `json:"priority" toml:"priority"` // 1 (highest) to 8 (lowest)
	Level            int             `json:"level" toml:"level"`       // Hierarchical level within requirements.
	User             string          `json:"user" toml:"user"`
	Status           string          `json:"status" toml:"status"` // e.g., "Draft", "Confirmed"
	CreatedAt        time.Time       `json:"created_at" toml:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" toml:"updated_at"`
	ParentID         int             `json:"parent_id" toml:"parent_id"` // 0 for top‑level
	Category         string          `json:"category" toml:"category"`   // e.g., "System Requirements"
	History          []ChangeLog     `json:"history" toml:"history"`     // Record of changes to the requirement.
	IntelligenceLink []*Intelligence `json:"intelligence_links" toml:"intelligence_links"`
	GateResults      []gates.Result  `json:"gate_results,omitempty" toml:"gate_results"`
	// Optional: Tags can help with flexible categorization or filtering.
	Tags []string `json:"tags,omitempty" toml:"tags"`
}

// Analyse sends the requirement description to the provided role/question pair
// and returns the result.
func (r *Requirement) Analyse(role, questionID string) (bool, string, error) {
	return interact.RunQuestion(gemini.DefaultClient, role, questionID, r.Description)
}

// EvaluateGates runs the specified gates against the requirement description
// and stores the results on the requirement.
func (r *Requirement) EvaluateGates(gateIDs []string) error {
	res, err := gates.Evaluate(gemini.DefaultClient, gateIDs, r.Description)
	if err != nil {
		return err
	}
	r.GateResults = res
	return nil
}

// SuggestOthers asks the client for related potential requirements based on
// this requirement's description and returns them.
func (r *Requirement) SuggestOthers(client gemini.Client) ([]Requirement, error) {
	prompt := fmt.Sprintf("Given the requirement %q, list other potential requirements (JSON array with `name` and `description`).", r.Description)
	resp, err := client.Ask(prompt)
	if err != nil {
		return nil, err
	}
	var reqs []Requirement
	if err := json.Unmarshal([]byte(resp), &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}

// FromGemini converts a Gemini requirement into a PMFS requirement.
// Only the ID, name, and description fields are mapped.
func FromGemini(req gemini.Requirement) Requirement {
	return Requirement{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
}

// Attachment is minimal metadata about an ingested file.
type Attachment struct {
	ID       int       `json:"id" toml:"id"`
	Filename string    `json:"filename" toml:"filename"`
	RelPath  string    `json:"rel_path" toml:"rel_path"` // e.g. "attachments/3/spec.pdf"
	Mimetype string    `json:"mimetype" toml:"mimetype"` // e.g. "application/pdf"
	AddedAt  time.Time `json:"added_at" toml:"added_at"`
	Analyzed bool      `json:"analyzed" toml:"analyzed"`
}

// Analyze processes the attachment with Gemini and appends proposed requirements.
func (att *Attachment) Analyze(prj *ProjectType) error {
	full := filepath.Join(projectDir(prj.ProductID, prj.ID), att.RelPath)
	reqs, err := gemini.AnalyzeAttachment(full)
	if err != nil {
		return err
	}
	for _, r := range reqs {
		prj.D.PotentialRequirements = append(prj.D.PotentialRequirements, FromGemini(r))
	}
	att.Analyzed = true
	return nil
}

// Analyse loads the attachment content and asks a role-specific question about it.
// For text files the content is read directly; for other files existing upload
// logic is used to extract textual content before querying the LLM.
func (att *Attachment) Analyse(role, questionID string, prj *ProjectType) (bool, string, error) {
	full := filepath.Join(projectDir(prj.ProductID, prj.ID), att.RelPath)
	mt := mime.TypeByExtension(strings.ToLower(filepath.Ext(full)))
	if i := strings.Index(mt, ";"); i >= 0 {
		mt = mt[:i]
	}
	var content string
	if strings.HasPrefix(mt, "text/") {
		b, err := os.ReadFile(full)
		if err != nil {
			return false, "", err
		}
		content = string(b)
	} else {
		reqs, err := gemini.AnalyzeAttachment(full)
		if err != nil {
			return false, "", err
		}
		var sb strings.Builder
		for _, r := range reqs {
			sb.WriteString(r.Name)
			sb.WriteString(": ")
			sb.WriteString(r.Description)
			sb.WriteString("\n")
		}
		content = sb.String()
	}
	return interact.RunQuestion(gemini.DefaultClient, role, questionID, content)
}

// ChangeLog records a change made to a requirement.
type ChangeLog struct {
	Timestamp time.Time `json:"timestamp" toml:"timestamp"`
	User      string    `json:"user" toml:"user"`
	Comment   string    `json:"comment" toml:"comment"`
}

// Intelligence represents data extracted from an attachment.
type Intelligence struct {
	ID int `json:"id" toml:"id"`
	// Source describes the type of attachment (e.g., "screenshot", "document").
	Filepath string `json:"Filepath" toml:"Filepath"`
	// Content contains the extracted text or metadata.
	Content     string    `json:"content" toml:"content"`
	Description string    `json:"description" toml:"description"`
	ExtractedAt time.Time `json:"extracted_at" toml:"extracted_at"`
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
		return fmt.Errorf("error saving index, AddProduct function: %w", err)
	}

	return nil
}

func (idx *Index) SaveIndex() error {
	// Project data is skipped automatically via struct tags, so the
	// index can be written directly without making a deep copy.
	if err := writeTOML(indexPath, idx); err != nil {
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
		return fmt.Errorf("error saving TOML, AddProject function: %w", err)
	}

	// Update in-memory index and persist
	prd.Projects = append(prd.Projects, addedproject)
	if err := idx.SaveIndex(); err != nil {
		return fmt.Errorf("error saving index, AddProject function: %w", err)
	}

	return nil
}

func (prj *ProjectType) SaveProject() error {
	prjDir := projectDir(prj.ProductID, prj.ID)
	if err := os.MkdirAll(prjDir, 0o755); err != nil {
		return fmt.Errorf("mkdir project dir: %w", err)
	}
	tomlPath := filepath.Join(prjDir, projectTOML)

	// Use a helper struct so ProjectData is included even though the
	// field is tagged with toml:"-" in ProjectType.
	type diskProject struct {
		ID        int         `toml:"id"`
		ProductID int         `toml:"productid"`
		Name      string      `toml:"name"`
		D         ProjectData `toml:"projectdata"`
	}

	dp := diskProject{
		ID:        prj.ID,
		ProductID: prj.ProductID,
		Name:      prj.Name,
		D:         prj.D,
	}

	if err := writeTOML(tomlPath, &dp); err != nil {
		return fmt.Errorf("error writing project TOML: %w", err)
	}
	return nil
}

// LoadProject loads a single project's TOML for this product.
func (prj *ProjectType) LoadProject() error {
	prjDir := projectDir(prj.ProductID, prj.ID)
	tomlPath := filepath.Join(prjDir, projectTOML)

	type diskProject struct {
		ID        int         `toml:"id"`
		ProductID int         `toml:"productid"`
		Name      string      `toml:"name"`
		D         ProjectData `toml:"projectdata"`
	}

	var dp diskProject
	if err := readTOML(tomlPath, &dp); err != nil {
		if os.IsNotExist(err) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("read project %s: %w", tomlPath, err)
	}

	prj.ID = dp.ID
	prj.ProductID = dp.ProductID
	prj.Name = dp.Name
	prj.D = dp.D
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

// moveFile tries rename, then falls back to copy+remove (cross-device safe).
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}

// detectMimeType reads the first 512 bytes to sniff the mimetype,
// falling back to extension-based detection.
func detectMimeType(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := io.ReadFull(f, buf)
	mt := http.DetectContentType(buf[:n])
	if mt == "application/octet-stream" {
		if ext := filepath.Ext(path); ext != "" {
			if byExt := mime.TypeByExtension(ext); byExt != "" {
				return byExt
			}
		}
	}
	return mt
}

// IngestInputDir scans inputDir and ingests all regular files into attachments/.
func (prj *ProjectType) IngestInputDir(inputDir string) ([]Attachment, error) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, err
	}

	// Optional: stable order
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if len(n) == 0 || n[0] == '.' {
			continue
		} // skip dotfiles
		names = append(names, n)
	}
	sort.Strings(names)

	ingested := make([]Attachment, 0, len(names))
	for _, n := range names {
		att, err := prj.AddAttachmentFromInput(inputDir, n)
		if err != nil {
			return ingested, err // fail fast; or change to continue if you prefer
		}
		ingested = append(ingested, att)
	}
	return ingested, nil
}

// AddAttachmentFromInput moves a single file from inputDir into this project's
// attachments/<id>/ folder, records minimal metadata, and saves the project.
func (prj *ProjectType) AddAttachmentFromInput(inputDir, filename string) (Attachment, error) {
	inputPath := filepath.Join(inputDir, filename)
	if ok, err := fileExists(inputPath); err != nil {
		return Attachment{}, err
	} else if !ok {
		return Attachment{}, fmt.Errorf("input file not found: %s", inputPath)
	}

	// Prepare destination base: .../projects/<prjID>/attachments/
	attBaseDir := attachmentDir(prj.ProductID, prj.ID)
	if err := os.MkdirAll(attBaseDir, 0o755); err != nil {
		return Attachment{}, fmt.Errorf("mkdir attachments: %w", err)
	}

	// Allocate next numeric ID using existing helper
	ids, err := numericSubdirs(attBaseDir)
	if err != nil {
		return Attachment{}, err
	}
	nextID := 1
	if len(ids) > 0 {
		nextID = ids[len(ids)-1] + 1
	}

	// Create the destination directory and final path
	attDir := filepath.Join(attBaseDir, strconv.Itoa(nextID))
	if err := os.MkdirAll(attDir, 0o755); err != nil {
		return Attachment{}, fmt.Errorf("mkdir attachment dir: %w", err)
	}
	base := filepath.Base(filename)
	dstPath := filepath.Join(attDir, base)

	// Move (rename with cross-device copy fallback)
	if err := moveFile(inputPath, dstPath); err != nil {
		return Attachment{}, fmt.Errorf("move file: %w", err)
	}

	// Detect mimetype
	mt := detectMimeType(dstPath)

	// Record attachment using a relative path for portability
	rel := filepath.ToSlash(filepath.Join("attachments", strconv.Itoa(nextID), base))
	att := Attachment{
		ID:       nextID,
		Filename: base,
		RelPath:  rel,
		Mimetype: mt,
		AddedAt:  time.Now(),
		Analyzed: false,
	}
	prj.D.Attachments = append(prj.D.Attachments, att)
	ptr := &prj.D.Attachments[len(prj.D.Attachments)-1]
	if err := ptr.Analyze(prj); err != nil {
		return *ptr, err
	}

	// Persist to project.toml
	if err := prj.SaveProject(); err != nil {
		return *ptr, err
	}
	return *ptr, nil
}
