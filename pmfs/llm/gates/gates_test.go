package gates

import (
	"fmt"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestGetGate(t *testing.T) {
	ids := []string{"clarity-form-1", "consistency-1", "duplicate-1"}
	for _, id := range ids {
		g, err := GetGate(id)
		if err != nil {
			t.Fatalf("GetGate(%s): %v", id, err)
		}
		if g.ID != id || g.Question == "" {
			t.Fatalf("unexpected gate for %s: %#v", id, g)
		}
	}
}

func TestGetGateUnknown(t *testing.T) {
	if _, err := GetGate("unknown"); err == nil {
		t.Fatalf("expected error for unknown gate")
	}
}

func TestEvaluate(t *testing.T) {
	text := "The system shall log in users"
	g1, _ := GetGate("clarity-form-1")
	g2, _ := GetGate("duplicate-1")
	expected1 := fmt.Sprintf("Given the requirement %s, %s Answer yes or no.", text, g1.Question)
	expected2 := fmt.Sprintf("Given the requirement %s, %s Answer yes or no.", text, g2.Question)
	expectedFollow := g2.FollowUp

	call := 0
	c := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		call++
		switch call {
		case 1:
			if prompt != expected1 {
				t.Fatalf("unexpected prompt1 %q", prompt)
			}
			return "Yes", nil
		case 2:
			if prompt != expected2 {
				t.Fatalf("unexpected prompt2 %q", prompt)
			}
			return "No", nil
		case 3:
			if prompt != expectedFollow {
				t.Fatalf("unexpected follow-up prompt %q", prompt)
			}
			return "Provide clarification", nil
		default:
			t.Fatalf("unexpected call %d with prompt %q", call, prompt)
			return "", nil
		}
	}}

	res, err := Evaluate(c, []string{"clarity-form-1", "duplicate-1"}, text)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if call != 3 {
		t.Fatalf("expected 3 Ask calls, got %d", call)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	m := map[string]Result{}
	for _, r := range res {
		m[r.Gate.ID] = r
	}
	if !m["clarity-form-1"].Pass {
		t.Fatalf("clarity-form-1 should pass")
	}
	if m["clarity-form-1"].FollowUp != "" {
		t.Fatalf("unexpected follow-up for clarity-form-1")
	}
	if m["duplicate-1"].Pass {
		t.Fatalf("duplicate-1 should fail")
	}
	if m["duplicate-1"].FollowUp != "Provide clarification" {
		t.Fatalf("unexpected follow-up %q", m["duplicate-1"].FollowUp)
	}
}

func TestEvaluateText(t *testing.T) {
	text := "The system shall log in users"
	g1, _ := GetGate("clarity-form-1")
	expected := fmt.Sprintf("Given the requirement %s, %s Answer yes or no.", text, g1.Question)

	call := 0
	c := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		call++
		if prompt != expected {
			t.Fatalf("unexpected prompt %q", prompt)
		}
		return "Yes", nil
	}}
	orig := gemini.SetClient(c)
	defer gemini.SetClient(orig)

	res, err := EvaluateText([]string{"clarity-form-1"}, text)
	if err != nil {
		t.Fatalf("EvaluateText: %v", err)
	}
	if len(res) != 1 || !res[0].Pass {
		t.Fatalf("unexpected result: %#v", res)
	}
	if call != 1 {
		t.Fatalf("expected 1 Ask call, got %d", call)
	}
}
