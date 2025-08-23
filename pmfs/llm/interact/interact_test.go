package interact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// createPromptFiles helper to create question and follow-up files.
func createPromptFiles(t *testing.T, dir, role, qID string) {
	t.Helper()
	rdir := filepath.Join(dir, role)
	if err := os.MkdirAll(rdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rdir, qID+".txt"), []byte("question"), 0o644); err != nil {
		t.Fatalf("write question: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rdir, qID+"_followup.txt"), []byte("follow"), 0o644); err != nil {
		t.Fatalf("write followup: %v", err)
	}
}

func TestRunQuestionResponseHandling(t *testing.T) {
	dir := t.TempDir()
	SetPromptsDir(dir)
	role, qID := "analyst", "Q1"
	createPromptFiles(t, dir, role, qID)

	tests := []struct {
		name        string
		responses   []string
		want        bool
		wantCalls   int
		checkFollow bool
	}{
		{"yes", []string{"Yes"}, true, 1, false},
		{"ambiguousThenYes", []string{"Maybe", "Yes"}, true, 2, false},
		{"noWithFollowUp", []string{"No", "ignored"}, false, 2, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var paths []string
			i := 0
			orig := gemini.SetClient(gemini.ClientFunc(func(path string) ([]gemini.Requirement, error) {
				paths = append(paths, path)
				resp := tc.responses[i]
				i++
				return []gemini.Requirement{{Description: resp}}, nil
			}))
			defer gemini.SetClient(orig)

			got, err := RunQuestion(role, qID)
			if err != nil {
				t.Fatalf("RunQuestion: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if len(paths) != tc.wantCalls {
				t.Fatalf("calls %d, want %d", len(paths), tc.wantCalls)
			}
			if tc.checkFollow {
				if !strings.HasSuffix(paths[1], qID+"_followup.txt") {
					t.Fatalf("expected follow-up call, got %v", paths)
				}
			}
		})
	}
}

func TestRunQuestionLoadsCorrectPrompt(t *testing.T) {
	dir := t.TempDir()
	SetPromptsDir(dir)

	cases := []struct{ role, qID string }{
		{"dev", "Q1"},
		{"tester", "Q2"},
	}

	for _, c := range cases {
		createPromptFiles(t, dir, c.role, c.qID)
		var gotPath string
		orig := gemini.SetClient(gemini.ClientFunc(func(path string) ([]gemini.Requirement, error) {
			gotPath = path
			return []gemini.Requirement{{Description: "Yes"}}, nil
		}))
		_, err := RunQuestion(c.role, c.qID)
		gemini.SetClient(orig)
		if err != nil {
			t.Fatalf("RunQuestion: %v", err)
		}
		expected := filepath.Join(dir, c.role, c.qID+".txt")
		if gotPath != expected {
			t.Errorf("for %s/%s, got path %s, want %s", c.role, c.qID, gotPath, expected)
		}
	}
}
