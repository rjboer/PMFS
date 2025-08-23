package prompts

import (
	"encoding/json"
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

// GetPrompts loads prompts from a file specified by the PMFS_PROMPTS_FILE
// environment variable. The file may be in JSON or TOML format and should map
// question identifiers to prompt templates.
func GetPrompts() (map[string]string, error) {
	path := os.Getenv("PMFS_PROMPTS_FILE")
	if path == "" {
		return nil, fmt.Errorf("PMFS_PROMPTS_FILE not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err == nil {
		return m, nil
	}
	if err := toml.Unmarshal(b, &m); err == nil {
		return m, nil
	}
	return nil, fmt.Errorf("unable to parse prompts file: %s", path)
}
