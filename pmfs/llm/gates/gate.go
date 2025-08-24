package gates

import (
	"fmt"

	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

// Gate represents a Yes/No gate used to evaluate requirements.
type Gate struct {
	ID       string
	Question string
	FollowUp string
}

var (
	registry    = map[string]Gate{}
	gatePrompts []prompts.Prompt
)

func register(g Gate) {
	registry[g.ID] = g
	template := fmt.Sprintf("Given the requirement %%s, %s Answer yes or no.", g.Question)
	gatePrompts = append(gatePrompts, prompts.Prompt{ID: g.ID, Template: template, FollowUp: g.FollowUp})
}

func init() {
	prompts.RegisterRole("quality_gate", gatePrompts)
}

// GetGate returns the gate with the given ID or an error if it doesn't exist.
func GetGate(id string) (Gate, error) {
	g, ok := registry[id]
	if !ok {
		return Gate{}, fmt.Errorf("gate %q not found", id)
	}
	return g, nil
}
