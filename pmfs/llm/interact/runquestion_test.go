package interact

import (
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

func TestRunQuestionYes(t *testing.T) {
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "1", Template: "Respond only with the word 'Yes'. Requirement: %s."}})
	calls := 0
	c := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		calls++
		expected := "Respond only with the word 'Yes'. Requirement: ignored."
		if prompt != expected {
			t.Fatalf("unexpected prompt %q", prompt)
		}
		return "Yes", nil
	}}

	got, follow, err := RunQuestion(c, "test", "1", "ignored")
	if err != nil {
		t.Fatalf("RunQuestion: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one Ask call, got %d", calls)
	}
	if !got {
		t.Fatalf("expected yes result")
	}
	if follow != "" {
		t.Fatalf("unexpected follow-up %q", follow)
	}
}

func TestRunQuestionNoFollowUp(t *testing.T) {
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "1", Template: "Respond only with the word 'No'. Requirement: %s.", FollowUp: "Reply with the word 'FollowUp'."}})
	call := 0
	c := gemini.ClientFunc{AskFunc: func(prompt string) (string, error) {
		call++
		switch call {
		case 1:
			expected := "Respond only with the word 'No'. Requirement: ignored."
			if prompt != expected {
				t.Fatalf("unexpected prompt %q", prompt)
			}
			return "No", nil
		case 2:
			expected := "Reply with the word 'FollowUp'."
			if prompt != expected {
				t.Fatalf("unexpected follow-up prompt %q", prompt)
			}
			return "FollowUp", nil
		default:
			t.Fatalf("unexpected call %d with prompt %q", call, prompt)
			return "", nil
		}
	}}

	got, follow, err := RunQuestion(c, "test", "1", "ignored")
	if err != nil {
		t.Fatalf("RunQuestion: %v", err)
	}
	if call != 2 {
		t.Fatalf("expected two Ask calls, got %d", call)
	}
	if got {
		t.Fatalf("expected false result")
	}
	if follow != "FollowUp" {
		t.Fatalf("unexpected follow-up %q", follow)
	}
}
