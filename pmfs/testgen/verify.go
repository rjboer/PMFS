package testgen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rjboer/PMFS/pmfs/llm/gemini"
)

const rulesTestTest1 = `Given the following Go code:

%s

and the specification:

%s

Does the code satisfy the specification? Respond with Yes or No and a short explanation.`

// Verify uses Gemini to determine whether code satisfies a specification.
// It returns true when the model indicates "Yes".
func Verify(code, spec string, c gemini.Client) (bool, error) {
	prompt := fmt.Sprintf(rulesTestTest1, code, spec)
	resp, err := c.Ask(prompt)
	if err != nil {
		return false, err
	}
	if ok, found := parseYesNo(resp); found {
		return ok, nil
	}

	follow := "Please answer only with 'Yes' or 'No': does the code satisfy the specification?"
	resp, err = c.Ask(follow)
	if err != nil {
		return false, err
	}
	if ok, found := parseYesNo(resp); found {
		return ok, nil
	}
	return false, fmt.Errorf("ambiguous response: %s", resp)
}

var yesNoRE = regexp.MustCompile(`(?i)\b(yes|no)\b`)

func parseYesNo(s string) (bool, bool) {
	m := yesNoRE.FindStringSubmatch(s)
	if len(m) != 2 {
		return false, false
	}
	return strings.EqualFold(m[1], "yes"), true
}
