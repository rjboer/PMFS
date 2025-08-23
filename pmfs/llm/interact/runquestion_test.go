package interact

import (
	"os"
	"strings"
	"testing"

	"github.com/rjboer/PMFS/pmfs/llm/prompts"
)

// loadAPIKey ensures GEMINI_API_KEY is available or skips the test.
func loadAPIKey(t *testing.T) {
	t.Helper()
	if _, ok := os.LookupEnv("GEMINI_API_KEY"); !ok {
		t.Skip("GEMINI_API_KEY not set")
	}
}

func TestRunQuestionYes(t *testing.T) {
	loadAPIKey(t)
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "1", Question: "Respond only with the word 'Yes'."}})
	got, follow, err := RunQuestion("test", "1")
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") {
			t.Skipf("API access forbidden: %v", err)
		}
		t.Fatalf("RunQuestion: %v", err)
	}
	if !got {
		t.Fatalf("expected yes result")
	}
	if follow != "" {
		t.Fatalf("unexpected follow-up %q", follow)
	}
}

func TestRunQuestionNoFollowUp(t *testing.T) {
	loadAPIKey(t)
	prompts.SetTestPrompts([]prompts.Prompt{{ID: "1", Question: "Respond only with the word 'No'.", FollowUp: "Reply with the word 'FollowUp'."}})
	got, follow, err := RunQuestion("test", "1")
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") {
			t.Skipf("API access forbidden: %v", err)
		}
		t.Fatalf("RunQuestion: %v", err)
	}
	if got {
		t.Fatalf("expected false result")
	}
	if !strings.Contains(strings.ToLower(follow), "followup") {
		t.Fatalf("unexpected follow-up %q", follow)
	}
}
