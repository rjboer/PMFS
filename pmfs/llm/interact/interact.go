package interact

import (
	"fmt"
	"regexp"
	"strings"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

var yesNoRE = regexp.MustCompile(`^(?:Yes|yes|No|no)$`)

// RunQuestion loads prompts, formats the selected question with text and asks
// Gemini for a Yes/No answer. The returned answer is normalized to "Yes" or
// "No".
func RunQuestion(role, questionID, text string) (string, error) {
	pr, err := prompts.GetPrompts()
	if err != nil {
		return "", err
	}
	tmpl, ok := pr[questionID]
	if !ok {
		return "", fmt.Errorf("unknown question id %s", questionID)
	}
	prompt := fmt.Sprintf(tmpl, text)
	answer, err := gemini.Client.Ask(role, prompt)
	if err != nil {
		return "", err
	}
	answer = strings.TrimSpace(answer)
	match := yesNoRE.FindString(answer)
	if match == "" {
		prompt = prompt + " Answer only 'Yes', 'yes', 'No', or 'no'"
		answer, err = gemini.Client.Ask(role, prompt)
		if err != nil {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		match = yesNoRE.FindString(answer)
		if match == "" {
			return "", fmt.Errorf("no yes/no answer returned")
		}
	}
	if strings.EqualFold(match, "yes") {
		return "Yes", nil
	}
	return "No", nil
}
