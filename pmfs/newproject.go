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

	// Respect runtime override of the base directory.
	if dir := os.Getenv("PMFS_BASEDIR"); dir != "" {
		PMFS.SetBaseDir(dir)
	}

	if err := PMFS.EnsureLayout(); err != nil {
		return nil, err
	}

	idx, err := PMFS.LoadIndex()
	if err != nil {
		return nil, err
	}

	if len(idx.Products) == 0 {
		if err := idx.AddProduct("Default Product"); err != nil {
			return nil, err
		}
	}

	prd := &idx.Products[0]
	if err := prd.AddProject(&idx, name); err != nil {
		return nil, err
	}
	return &prd.Projects[len(prd.Projects)-1], nil
}
