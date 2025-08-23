package gates

import "fmt"

// Gate represents a Yes/No gate used to evaluate requirements.
type Gate struct {
	ID       string
	Question string
	FollowUp string
}

var registry = map[string]Gate{}

func register(g Gate) { registry[g.ID] = g }

// GetGate returns the gate with the given ID or an error if it doesn't exist.
func GetGate(id string) (Gate, error) {
	g, ok := registry[id]
	if !ok {
		return Gate{}, fmt.Errorf("gate %q not found", id)
	}
	return g, nil
}
