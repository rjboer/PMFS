package interact

import (
	"errors"
	"fmt"
	"regexp"
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

	prompt := fmt.Sprintf(p.Template, text)
	resp, err := client.Ask(prompt)
	if err != nil {
		return false, "", err
	}
	re := regexp.MustCompile(`(?i)\b(yes|no)\b`)
	match := re.FindStringSubmatch(resp)
	for i := 0; i < 2 && len(match) == 0; i++ {
		resp, err = client.Ask("Answer Yes or No only")
		if err != nil {
			return false, "", err
		}
		match = re.FindStringSubmatch(resp)
	}
	if len(match) == 0 {
		return false, "", errors.New("unable to determine yes/no answer")
	}
	ans := strings.ToLower(match[1])
	switch ans {
	case "yes":
		return true, "", nil
	case "no":
		if p.FollowUp == "" {
			return false, "", nil
		}
		follow, err := client.Ask(p.FollowUp)
		if err != nil {
			return false, "", err
		}
		return false, follow, nil
	}
	return false, "", errors.New("unexpected answer")
}
