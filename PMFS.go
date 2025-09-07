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
	llm "github.com/rjboer/PMFS/pmfs/llm"
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

// Database holds the products and the base directory from which it was loaded.
// The BaseDir field is not persisted to disk.
type Database struct {
	BaseDir  string        `toml:"-"`
	Products []ProductType `toml:"products"`
	LLM      llm.Client    `toml:"-" json:"-"`
}

// DB is the package-wide database instance used by helper functions.
var DB *Database

// DesignAspectGateGroup lists gate IDs evaluated for design aspect templates.
var DesignAspectGateGroup = []string{
	"clarity-form-1",
	"duplicate-1",
}

// LoadSetup initialises the database at the provided path. It sets the
// PMFS_BASEDIR environment variable, prepares the on-disk layout and loads the
// index into memory.
func LoadSetup(path string) (*Database, error) {
	// Export base directory for any helpers relying on the environment
	// variable and update internal path bookkeeping.
	if err := os.Setenv(envBaseDir, path); err != nil {
		return nil, err
	}
	SetBaseDir(path)

	if err := ensureLayout(); err != nil {
		return nil, err
	}
	db, err := loadDatabase()
	if err != nil {
		return nil, err
	}
	db.BaseDir = path
	db.LLM = llm.DefaultClient
	DB = db

	return db, nil
}

// Save persists the in-memory database back to disk.
func (db *Database) Save() error {
	return writeTOML(indexPath, db)
}

// -----------------------------------------------------------------------------
// Memory model
// -----------------------------------------------------------------------------

// ProductType represents a product within the database. Next IDs are derived
// from len(db.Products)+1.
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
	Name         string        `json:"name" toml:"name"`
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
	// RequirementRelations holds the LLM-scored relationships between requirements.
	//	RequirementRelations  []RequirementRelation `json:"requirement_relations" toml:"requirement_relations"`
	//	RequirementCategories []string              `json:"requirement_Categories" toml:"requirement_Categories"`
	FixedCategories bool `json:"requirement_FixedCategories" toml:"requirement_FixedCategories"`
}

// ConditionType represents the state of a requirement.
type ConditionType struct {
	Proposed    bool `json:"proposed" toml:"proposed"`
	AIgenerated bool `json:"aigenerated" toml:"aigenerated"`
	Active      bool `json:"active" toml:"active"`
	Deleted     bool `json:"deleted" toml:"deleted"`
}

// Requirement represents a confirmed requirement with detailed metadata.
type Requirement struct {
	ID                 int             `json:"id" toml:"id"`
	Name               string          `json:"name" toml:"name"`
	Description        string          `json:"description" toml:"description"`
	Priority           int             `json:"priority" toml:"priority"` // 1 (highest) to 8 (lowest)
	Level              int             `json:"level" toml:"level"`       // Hierarchical level within requirements.
	User               string          `json:"user" toml:"user"`
	Status             string          `json:"status" toml:"status"` // e.g., "Draft", "Confirmed"
	CreatedAt          time.Time       `json:"created_at" toml:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" toml:"updated_at"`
	ParentID           int             `json:"parent_id" toml:"parent_id"` // index of originating requirement
	AttachmentIndex    int             `json:"attachment_index" toml:"attachment_index"`
	Category           string          `json:"category" toml:"category"` // e.g., "System Requirements"
	History            []ChangeLog     `json:"history" toml:"history"`   // Record of changes to the requirement.
	IntelligenceLink   []*Intelligence `json:"intelligence_links" toml:"intelligence_links"`
	GateResults        []gates.Result  `json:"gate_results,omitempty" toml:"gate_results"`
	RecommendedChanges []DesignAspect  `json:"RequirementImprovements" toml:"requirementdesignaspects"`
	DesignAspects      []DesignAspect  `json:"design_aspects" toml:"design_aspects"`
	Condition          ConditionType   `json:"condition" toml:"condition"`
	// Optional: Tags can help with flexible categorization or filtering.
	Tags []string `json:"tags,omitempty" toml:"tags"`
}

// A DesignAspect is a take on the requirement, as a way to improve this,  as with the following example:
// DesignAspects describe an element of the project -> this turns into requirements.
// The following (bad dual requirement)
// The logistics line should have transport belts, these belts should be 4 meters long
// Improvements:
// Belt Material
// Belt Width
// Belt Speed
// Belt Capacity
// Belt Drive Mechanism
// Belt Cleaning Mechanism
// Emergency Stop Mechanism
// Belt Alignment System
// Belt Tensioning System
// Number of Belts
// There are several ways to improve the requirement or add new one
type DesignAspect struct {
	Name        string        `json:"name" toml:"name"`
	Description string        `json:"description" toml:"description"`
	Templates   []Requirement `json:"templates" toml:"templates"`
	Processed   bool          `json:"AspectProcessed" toml:"AspectProcessed"` //the aspect has been processed
}

// FromGemini converts a Gemini requirement into a PMFS requirement.
func FromGemini(req gemini.Requirement) Requirement {
	return Requirement{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
}

// Analyse sends the requirement description to the provided role/question pair

// using the database's configured LLM and returns the result.
func (r *Requirement) Analyse(role, questionID string) (bool, string, error) {
	return interact.RunQuestion(DB.LLM, role, questionID, r.Description)
}

// EvaluateGates runs the specified gates against the requirement description
// using the database's configured LLM and stores the results on the requirement.
func (r *Requirement) EvaluateGates(gateIDs []string) error {
	res, err := gates.Evaluate(DB.LLM, gateIDs, r.Description)

	if err != nil {
		return err
	}
	r.GateResults = res
	return nil
}

// QualityControlAI runs Analyse and EvaluateGates on the requirement.
// It returns the result of Analyse and stores gate evaluation results on the requirement.
func (r *Requirement) QualityControlAI(role, questionID string, gateIDs []string) (bool, string, error) {
	pass, ans, err := r.Analyse(role, questionID)
	if err != nil {
		return pass, ans, err
	}
	if err := r.EvaluateGates(gateIDs); err != nil {
		return pass, ans, err
	}
	return pass, ans, nil
}

// EvaluateDesignGates runs the specified gates against each template requirement
// in the design aspect using the database's configured LLM.
func (da *DesignAspect) EvaluateDesignGates(gateIDs []string) error {
	for i := range da.Templates {
		if err := da.Templates[i].EvaluateGates(gateIDs); err != nil {
			return err
		}
	}
	return nil
}

// Deduplicate removes or merges near-identical requirements using the configured LLM
// for semantic similarity. If the LLM is unavailable, a simple case-insensitive
// comparison of names and descriptions is used. Requirements marked as deleted are
// skipped entirely. If ignoreProposed is true, requirements marked as proposed are
// also skipped during duplicate comparison (but are still returned).
func Deduplicate(reqs []Requirement, ignoreProposed bool) []Requirement {
	var out []Requirement
	for _, r := range reqs {
		if r.Condition.Deleted {
			continue
		}
		if ignoreProposed && r.Condition.Proposed {
			out = append(out, r)
			continue
		}
		merged := false
		for i := range out {
			if ignoreProposed && out[i].Condition.Proposed {
				continue
			}
			same := false
			if DB != nil && DB.LLM != nil {
				prompt := fmt.Sprintf("Are the following two requirements essentially the same? Respond with 'yes' or 'no'.\n1. %s\n2. %s", out[i].Description, r.Description)
				if resp, err := DB.LLM.Ask(prompt); err == nil {
					resp = strings.ToLower(strings.TrimSpace(resp))
					if strings.HasPrefix(resp, "yes") {
						same = true
					}
				}
			} else {
				same = strings.EqualFold(out[i].Description, r.Description) || strings.EqualFold(out[i].Name, r.Name)
			}
			if same {
				if out[i].Description == "" && r.Description != "" {
					out[i].Description = r.Description
				}
				if out[i].Name == "" && r.Name != "" {
					out[i].Name = r.Name
				}
				merged = true
				break
			}
		}
		if !merged {
			out = append(out, r)
		}
	}
	return out
}

// parseLLMJSON extracts the first valid JSON array from the LLM response.
// It supports Markdown fenced code blocks and returns an error if no JSON
// array can be located.
func parseLLMJSON(resp string) ([]byte, error) {
	resp = strings.TrimSpace(resp)

	if idx := strings.Index(resp, "```"); idx != -1 {
		resp = resp[idx+3:]
		if nl := strings.IndexByte(resp, '\n'); nl != -1 {
			resp = resp[nl+1:]
		}
		if end := strings.Index(resp, "```"); end != -1 {
			resp = resp[:end]
		}
		resp = strings.TrimSpace(resp)
	}

	if json.Valid([]byte(resp)) {
		return []byte(resp), nil
	}

	start := strings.Index(resp, "[")
	for start != -1 {
		depth := 0
		for i := start; i < len(resp); i++ {
			switch resp[i] {
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					candidate := resp[start : i+1]
					if json.Valid([]byte(candidate)) {
						return []byte(candidate), nil
					}
					break
				}
			}
		}
		next := strings.Index(resp[start+1:], "[")
		if next == -1 {
			break
		}
		start += next + 1
	}
	return nil, errors.New("no valid JSON array found in LLM response")
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

// Analyze processes the attachment using the default strategy and appends
// proposed requirements. It is kept for backward compatibility and delegates to
// GenerateRequirements with an empty strategy.
func (att *Attachment) Analyze(prj *ProjectType) error {
	return att.GenerateRequirements(prj, "")
}

// GenerateRequirements analyzes the attachment using the provided heuristic
// strategy and appends any discovered requirements to the project's requirement
// slice. An empty strategy falls back to the default LLM-based
// analysis (currently Gemini).
func (att *Attachment) GenerateRequirements(prj *ProjectType, strategy string) error {
	if strategy == "" {
		strategy = "gemini"
	}

	full := filepath.Join(projectDir(prj.ProductID, prj.ID), att.RelPath)

	reqs, err := DB.LLM.AnalyzeAttachment(full)
	if err != nil {
		return err
	}
	attIdx := -1
	for i := range prj.D.Attachments {
		if &prj.D.Attachments[i] == att {
			attIdx = i
			break
		}
	}
	var newReqs []Requirement
	for _, r := range reqs {
		nr := FromGemini(r)
		nr.AttachmentIndex = attIdx
		nr.Condition.Proposed = true
		nr.Condition.AIgenerated = true
		newReqs = append(newReqs, nr)
	}
	prj.D.Requirements = Deduplicate(append(prj.D.Requirements, newReqs...), false)
	att.Analyzed = true

	// Summarize attachment content into an Intelligence entry.
	mt := mime.TypeByExtension(strings.ToLower(filepath.Ext(full)))
	if i := strings.Index(mt, ";"); i >= 0 {
		mt = mt[:i]
	}
	var content string
	if strings.HasPrefix(mt, "text/") {
		b, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		content = string(b)
	} else {
		var sb strings.Builder
		for _, r := range reqs {
			sb.WriteString(r.Name)
			sb.WriteString(": ")
			sb.WriteString(r.Description)
			sb.WriteString("\n")
		}
		content = sb.String()
	}
	summary, err := summarizeContent(content)
	if err != nil {
		return err
	}
	intel := Intelligence{
		ID:          len(prj.D.Intelligence) + 1,
		Filepath:    att.RelPath,
		Content:     content,
		Description: summary,
		ExtractedAt: time.Now(),
	}

	aspects, err := designAspectsFromSummary(summary)
	if err != nil {
		return err
	}
	intel.DesignAngles = append(intel.DesignAngles, aspects...)
	prj.D.Intelligence = append(prj.D.Intelligence, intel)

	// Persist newly added potential requirements and intelligence immediately.
	if err := prj.Save(); err != nil {
		return err
	}
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
		reqs, err := DB.LLM.AnalyzeAttachment(full)
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
	return interact.RunQuestion(DB.LLM, role, questionID, content)
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
	//DesignAngles describe ways/topics to describe the functionality that is caputured in the intelligence.
	DesignAngles []DesignAspect `json:"DesignAngles_DesignAspects" toml:"DesignAngles_DesignAspects"`
}

// -----------------------------------------------------------------------------
// Init helpers
// -----------------------------------------------------------------------------
//
// ensureLayout creates base folder structure and ensures index.toml exists.
func ensureLayout() error {
	if err := os.MkdirAll(baseProductsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", baseProductsDir, err)
	}
	// Create empty index if missing.
	if ok, err := fileExists(indexPath); err != nil {
		return err
	} else if !ok {
		db := Database{Products: []ProductType{}}
		if err := writeTOML(indexPath, &db); err != nil {
			return fmt.Errorf("write index.toml: %w", err)
		}
	}
	return nil
}

// loadDatabase reads index.toml into the database model.
func loadDatabase() (*Database, error) {
	var db Database
	if err := readTOML(indexPath, &db); err != nil {
		if os.IsNotExist(err) {
			// Create a fresh one if missing (keeps flow simple)
			db = Database{Products: []ProductType{}}
			if werr := writeTOML(indexPath, &db); werr != nil {
				return nil, werr
			}
			return &db, nil
		}
		return nil, fmt.Errorf("read index.toml: %w", err)
	}
	if db.Products == nil {
		db.Products = []ProductType{}
	}
	return &db, nil
}

// -----------------------------------------------------------------------------
// Public ops
// -----------------------------------------------------------------------------

// ProductData holds metadata for products persisted in the index.
type ProductData struct {
	ID   int    `json:"id" toml:"id"`
	Name string `json:"name" toml:"name"`
}

// NewProduct creates a product, persists it to index.toml and returns its ID.
func (db *Database) NewProduct(data ProductData) (int, error) {
	if strings.TrimSpace(data.Name) == "" {
		return 0, errors.New("product name cannot be empty")
	}

	newID := len(db.Products) + 1
	pDir := productDir(newID)
	if err := os.MkdirAll(filepath.Join(pDir, "projects"), 0o755); err != nil {

		return 0, fmt.Errorf("mkdir product/projects: %w", err)
	}
	prd := ProductType{ID: newID, Name: data.Name, Projects: []ProjectType{}}
	db.Products = append(db.Products, prd)
	if err := db.Save(); err != nil {
		return 0, err
	}
	return newID, nil
}

// ModifyProduct updates product fields and persists the index.
func (db *Database) ModifyProduct(data ProductData) (int, error) {
	for i := range db.Products {
		if db.Products[i].ID == data.ID {
			if data.Name != "" {
				db.Products[i].Name = data.Name
			}
			if err := db.Save(); err != nil {
				return 0, err
			}
			return data.ID, nil
		}
	}
	return 0, ErrProductNotFound
}

// NewProject appends a project to the product, persists it to project.toml and
// updates the database index. The new project ID is len(product.Projects)+1.
func (prd *ProductType) NewProject(data ProjectData) (int, error) {
	// basic validation
	if strings.TrimSpace(data.Name) == "" {
		return 0, errors.New("project name cannot be empty")
	}

	newPrjID := len(prd.Projects) + 1
	prjDir := projectDir(prd.ID, newPrjID)

	// ensure directory exists
	if err := os.MkdirAll(prjDir, 0o755); err != nil {
		return 0, fmt.Errorf("mkdir project dir: %w", err)
	}

	prj := ProjectType{
		ID:        newPrjID,
		ProductID: prd.ID,
		Name:      data.Name,
		D:         data,
	}

	if err := prj.Save(); err != nil {
		return 0, fmt.Errorf("error saving TOML, NewProject function: %w", err)
	}

	prd.Projects = append(prd.Projects, prj)
	if err := DB.Save(); err != nil {
		return 0, err
	}
	return newPrjID, nil
}

// ModifyProject updates a project's fields and persists both the project and index.
func (prd *ProductType) ModifyProject(id int, data ProjectData) (int, error) {
	for i := range prd.Projects {
		if prd.Projects[i].ID == id {
			if data.Name != "" {
				prd.Projects[i].Name = data.Name
				prd.Projects[i].D.Name = data.Name
			} else {
				data.Name = prd.Projects[i].D.Name
			}
			prd.Projects[i].D = data

			if err := prd.Projects[i].Save(); err != nil {
				return 0, err
			}
			if err := DB.Save(); err != nil {
				return 0, err
			}
			return id, nil
		}
	}
	return 0, ErrProjectNotFound
}

// Save writes the project's data to its project.toml.
func (prj *ProjectType) Save() error {
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

// Load loads a single project's TOML for this product.
func (prj *ProjectType) Load() error {
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

// Project returns the project with the given ID for this product by loading it
// from its on-disk TOML. If the project is not listed in the product, an
// ErrProjectNotFound is returned.
func (prd *ProductType) Project(id int) (*ProjectType, error) {
	for i := range prd.Projects {
		if prd.Projects[i].ID == id {
			prd.Projects[i].ProductID = prd.ID
			if err := prd.Projects[i].Load(); err != nil {
				return nil, err
			}
			return &prd.Projects[i], nil
		}
	}
	return nil, ErrProjectNotFound
}

// LoadProjects loads all listed projects for this product by reading each project.toml.
func (prd *ProductType) LoadProjects() error {
	if prd.Projects == nil || len(prd.Projects) == 0 {
		return nil
	}
	for i := range prd.Projects {
		prd.Projects[i].ProductID = prd.ID
		if err := prd.Projects[i].Load(); err != nil {
			return err
		}
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
	if err := prj.Save(); err != nil {
		return *ptr, err
	}
	return *ptr, nil
}

// AddAttachmentFromText creates a new attachment from the provided text
// content and analyzes it using the configured LLM.
func (prj *ProjectType) AddAttachmentFromText(text string) (Attachment, error) {
	attBaseDir := attachmentDir(prj.ProductID, prj.ID)
	if err := os.MkdirAll(attBaseDir, 0o755); err != nil {
		return Attachment{}, fmt.Errorf("mkdir attachments: %w", err)
	}

	ids, err := numericSubdirs(attBaseDir)
	if err != nil {
		return Attachment{}, err
	}
	nextID := 1
	if len(ids) > 0 {
		nextID = ids[len(ids)-1] + 1
	}

	attDir := filepath.Join(attBaseDir, strconv.Itoa(nextID))
	if err := os.MkdirAll(attDir, 0o755); err != nil {
		return Attachment{}, fmt.Errorf("mkdir attachment dir: %w", err)
	}
	filename := "note.txt"
	dstPath := filepath.Join(attDir, filename)
	if err := os.WriteFile(dstPath, []byte(text), 0o644); err != nil {
		return Attachment{}, err
	}

	rel := filepath.ToSlash(filepath.Join("attachments", strconv.Itoa(nextID), filename))
	att := Attachment{
		ID:       nextID,
		Filename: filename,
		RelPath:  rel,
		Mimetype: "text/plain",
		AddedAt:  time.Now(),
		Analyzed: false,
	}
	prj.D.Attachments = append(prj.D.Attachments, att)
	ptr := &prj.D.Attachments[len(prj.D.Attachments)-1]
	if err := ptr.Analyze(prj); err != nil {
		return *ptr, err
	}
	if err := prj.Save(); err != nil {
		return *ptr, err
	}
	return *ptr, nil
}

// ActivateRequirementByID activates the requirement with the given ID.
// It sets Proposed to false and Active to true.
func (prj *ProjectType) ActivateRequirementByID(id int) {
	for i := range prj.D.Requirements {
		if prj.D.Requirements[i].ID == id {
			prj.D.Requirements[i].Condition.Proposed = false
			prj.D.Requirements[i].Condition.Active = true
			_ = prj.Save()
			return
		}
	}
}

// ActivateRequirementsWhere activates all requirements for which pred returns true.
// It toggles Proposed to false and Active to true for matches.
func (prj *ProjectType) ActivateRequirementsWhere(pred func(Requirement) bool) {
	for i := range prj.D.Requirements {
		if pred(prj.D.Requirements[i]) {
			prj.D.Requirements[i].Condition.Proposed = false
			prj.D.Requirements[i].Condition.Active = true
		}
	}
	_ = prj.Save()
}

// AddRequirement appends a requirement to the project and persists it.
func (prj *ProjectType) AddRequirement(r Requirement) error {
	r.ID = len(prj.D.Requirements) + 1
	prj.D.Requirements = append(prj.D.Requirements, r)
	return prj.Save()
}

// GenerateDesignAspectsAll runs GenerateDesignAspects for every requirement in
// the project and persists the results.
func (prj *ProjectType) GenerateDesignAspectsAll() error {
	for i := range prj.D.Requirements {
		if _, err := prj.D.Requirements[i].GenerateDesignAspects(); err != nil {
			return err
		}
	}
	return prj.Save()
}

// QualityControlScanALL runs QualityControlAI on every requirement in the project.
func (prj *ProjectType) QualityControlScanALL(role, questionID string, gateIDs []string) error {
	for i := range prj.D.Requirements {
		if _, _, err := prj.D.Requirements[i].QualityControlAI(role, questionID, gateIDs); err != nil {
			return err
		}
	}
	return nil
}

// AnalyseAll runs QualityControlAI on all requirements.
// It processes all requirements, returning the first error encountered, and
// persists any gate evaluation results.
func (prj *ProjectType) AnalyseAll(role, questionID string, gateIDs []string) error {
	var firstErr error

	for i := range prj.D.Requirements {
		req := &prj.D.Requirements[i]
		if req.Condition.Proposed || req.Condition.Deleted {
			continue
		}
		if _, _, err := req.QualityControlAI(role, questionID, gateIDs); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if err := prj.Save(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}
