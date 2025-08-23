package interact

import (
	"errors"
	"fmt"
	"strings"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

// RunQuestion formats the question template for a role with the provided text
// and asks it using the supplied Gemini client. It returns true when the response
// contains "yes". When the response contains "no" and the prompt defines a
// follow-up question, the follow-up is sent and its response returned alongside
// the false result.
func RunQuestion(client gemini.Client, role, questionID, text string) (bool, string, error) {
	ps, err := prompts.GetPrompts(role)
	if err != nil {
		return false, "", err
	}
	var p *prompts.Prompt
	for i := range ps {
		if ps[i].ID == questionID {
			p = &ps[i]
			break
		}
	}
	if p == nil {
		return false, "", fmt.Errorf("prompt %s/%s not found", role, questionID)
	}

	for i := 0; i < 3; i++ {
		prompt := fmt.Sprintf(p.Template, text)
		resp, err := client.Ask(prompt)
		if err != nil {
			return false, "", err
		}
		ans := strings.ToLower(resp)
		switch {
		case strings.Contains(ans, "yes"):
			return true, "", nil
		case strings.Contains(ans, "no"):
			if p.FollowUp == "" {
				return false, "", nil
			}
			follow, err := client.Ask(p.FollowUp)
			if err != nil {
				return false, "", err
			}
			return false, follow, nil
		}
	}
	return false, "", errors.New("unable to determine answer")
}
