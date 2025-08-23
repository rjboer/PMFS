package testgen

import (
	"testing"

	"github.com/rjboer/PMFS/pmfs/llm/gemini"
)

func TestVerifyYes(t *testing.T) {
	var calls int
	c := gemini.ClientFunc{
		AskFunc: func(prompt string) (string, error) {
			calls++
			return "Yes, it works", nil
		},
	}
	ok, err := Verify("code", "spec", c)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestVerifyNo(t *testing.T) {
	c := gemini.ClientFunc{
		AskFunc: func(prompt string) (string, error) {
			return "No, it fails", nil
		},
	}
	ok, err := Verify("code", "spec", c)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected false, got true")
	}
}

func TestVerifyAmbiguous(t *testing.T) {
	var prompts []string
	c := gemini.ClientFunc{
		AskFunc: func(prompt string) (string, error) {
			prompts = append(prompts, prompt)
			if len(prompts) == 1 {
				return "It's hard to tell", nil
			}
			return "Yes", nil
		},
	}
	ok, err := Verify("code", "spec", c)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected true, got false")
	}
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}
}
