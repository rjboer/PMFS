package gates

import (
	llm "github.com/rjboer/PMFS/pmfs/llm"
	"github.com/rjboer/PMFS/pmfs/llm/interact"
)

// Result holds the outcome of a gate evaluation.
type Result struct {
	Gate     Gate
	Pass     bool
	FollowUp string
}

// Evaluate runs the specified gates against the provided text using the LLM client.
// It returns a Result for each gate in the same order as gateIDs.
func Evaluate(client llm.Client, gateIDs []string, text string) ([]Result, error) {
	var results []Result
	for _, id := range gateIDs {
		g, err := GetGate(id)
		if err != nil {
			return nil, err
		}
		pass, follow, err := interact.RunQuestion(client, "quality_gate", g.ID, text)
		if err != nil {
			return nil, err
		}
		results = append(results, Result{Gate: g, Pass: pass, FollowUp: follow})
	}
	return results, nil
}
