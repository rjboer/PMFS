package gates

import (
	"fmt"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/interact"
	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

// Result holds the outcome of a gate evaluation.
type Result struct {
	Gate     Gate
	Pass     bool
	FollowUp string
}

// Evaluate runs the specified gates against the provided text using the Gemini client.
// It returns a Result for each gate in the same order as gateIDs.
func Evaluate(client gemini.Client, gateIDs []string, text string) ([]Result, error) {
	var results []Result
	for _, id := range gateIDs {
		g, err := GetGate(id)
		if err != nil {
			return nil, err
		}
		template := fmt.Sprintf("Given the requirement %%s, %s Answer yes or no.", g.Question)
		prompts.SetTestPrompts([]prompts.Prompt{{ID: g.ID, Template: template, FollowUp: g.FollowUp}})
		pass, follow, err := interact.RunQuestion(client, "test", g.ID, text)
		if err != nil {
			return nil, err
		}
		results = append(results, Result{Gate: g, Pass: pass, FollowUp: follow})
	}
	return results, nil
}
