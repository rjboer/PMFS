//go:build test

package interact

import (
	"errors"
	"path/filepath"
	"strings"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

var promptsDir string

// SetPromptsDir overrides the directory used to load prompts.
func SetPromptsDir(dir string) { promptsDir = dir }

// RunQuestion is a lightweight stub used only in tests when the real
// implementation is unavailable. It sends the prompt to the configured
// Gemini client and interprets the first requirement description as a
// yes/no response. Ambiguous answers are retried up to three times.
// A follow-up prompt is sent when the response is "No".
func RunQuestion(role, questionID string) (bool, error) {
	if promptsDir == "" {
		return false, errors.New("prompts directory not set")
	}
	questionPath := filepath.Join(promptsDir, role, questionID+".txt")
	followupPath := filepath.Join(promptsDir, role, questionID+"_followup.txt")
	for i := 0; i < 3; i++ {
		reqs, err := gemini.AnalyzeAttachment(questionPath)
		if err != nil {
			return false, err
		}
		if len(reqs) == 0 {
			continue
		}
		ans := strings.ToLower(reqs[0].Description)
		switch {
		case strings.Contains(ans, "yes"):
			return true, nil
		case strings.Contains(ans, "no"):
			// send follow-up but ignore response
			gemini.AnalyzeAttachment(followupPath)
			return false, nil
		}
	}
	return false, errors.New("unable to determine answer")
}
