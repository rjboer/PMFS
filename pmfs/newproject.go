package pmfs

import (
	"os"

	PMFS "github.com/rjboer/PMFS"
	"github.com/rjboer/PMFS/pmfs/llm"
	"github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// ProjectType is an alias to the core PMFS project type.
type ProjectType = PMFS.ProjectType

// NewProject ensures the data layout exists, assigns the default LLM client
// from the environment, and creates a new project with the provided name under
// the first product (creating a default product if necessary).
func NewProject(name string) (*ProjectType, error) {
	// Ensure the default client uses the API key from the environment.
	llm.SetClient(gemini.NewRESTClient(os.Getenv("GEMINI_API_KEY")))

	dir := os.Getenv("PMFS_BASEDIR")
	if dir == "" {
		dir = "database"
	}

	db, err := PMFS.LoadSetup(dir)
	if err != nil {
		return nil, err
	}

	if len(db.Products) == 0 {

		if _, err := db.NewProduct(PMFS.ProductData{Name: "Default Product"}); err != nil {

			return nil, err
		}
	}

	prd := &db.Products[0]

	prj, err := prd.NewProject(name)

	if err != nil {
		return nil, err
	}
	if err := db.Save(); err != nil {
		return nil, err
	}
	return prj, nil
}
