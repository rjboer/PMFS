package gemini_test

import (
	"testing"

	"github.com/rjboer/PMFS/pmfs/llm"
	gemini "github.com/rjboer/PMFS/pmfs/llm/gemini"
)

type testClient struct{ name string }

func (t testClient) AnalyzeAttachment(path string) ([]gemini.Requirement, error) {
	return []gemini.Requirement{{Name: t.name}}, nil
}

func (t testClient) Ask(prompt string) (string, error) {
	return t.name, nil
}

func TestSetClientSwapsImplementation(t *testing.T) {
	c1 := testClient{name: "first"}
	c2 := testClient{name: "second"}

	prev := llm.SetClient(c1)
	t.Cleanup(func() { llm.SetClient(prev) })

	reqs, err := llm.AnalyzeAttachment("p")
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].Name != "first" {
		t.Fatalf("unexpected requirements from c1: %#v", reqs)
	}

	prev2 := llm.SetClient(c2)
	if prev2 != c1 {
		t.Fatalf("expected previous client to be c1")
	}

	reqs, err = llm.AnalyzeAttachment("p")
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].Name != "second" {
		t.Fatalf("unexpected requirements from c2: %#v", reqs)
	}
}
