package gemini

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRESTClientInitUsesAPIKeyFromEnv(t *testing.T) {
	key := "test-key"
	if b, err := os.ReadFile(filepath.Join("..", "..", "..", ".env")); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "GEMINI_API_KEY=") {
				key = strings.TrimSpace(strings.TrimPrefix(line, "GEMINI_API_KEY="))
				break
			}
		}
	}
	t.Setenv("GEMINI_API_KEY", key)
	c := &RESTClient{}
	if err := c.init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if c.APIKey != key {
		t.Fatalf("expected APIKey %q, got %q", key, c.APIKey)
	}
}

type mockTransport struct{}

func (mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "/upload/") {
		rec := httptest.NewRecorder()
		rec.WriteString(`{"file":{"name":"files/1","mimeType":"text/plain"}}`)
		return rec.Result(), nil
	}
	if strings.Contains(req.URL.Path, ":generateContent") {
		rec := httptest.NewRecorder()
		body := map[string]any{
			"candidates": []any{map[string]any{
				"content": map[string]any{
					"parts": []any{map[string]any{"text": `[{"id":1,"name":"R1","description":"D1"}]`}},
				},
			}},
		}
		b, _ := json.Marshal(body)
		rec.Write(b)
		return rec.Result(), nil
	}
	return nil, fmt.Errorf("unexpected path %s", req.URL.Path)
}

func TestRESTClientAnalyzeAttachment(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(fpath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	c := &RESTClient{APIKey: "k", HTTPClient: &http.Client{Transport: mockTransport{}}}
	reqs, err := c.AnalyzeAttachment(fpath)
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].Name != "R1" {
		t.Fatalf("unexpected requirements: %#v", reqs)
	}
}

type testClient struct{ name string }

func (t testClient) AnalyzeAttachment(path string) ([]Requirement, error) {
	return []Requirement{{Name: t.name}}, nil
}

func TestSetClientSwapsImplementation(t *testing.T) {
	c1 := testClient{name: "first"}
	c2 := testClient{name: "second"}

	prev := SetClient(c1)
	defer SetClient(prev)

	reqs, err := AnalyzeAttachment("p")
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].Name != "first" {
		t.Fatalf("unexpected requirements from c1: %#v", reqs)
	}

	prev2 := SetClient(c2)
	if prev2 != c1 {
		t.Fatalf("expected previous client to be c1")
	}
	reqs, err = AnalyzeAttachment("p")
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].Name != "second" {
		t.Fatalf("unexpected requirements from c2: %#v", reqs)
	}
}

func TestClientFuncAnalyzeAttachment(t *testing.T) {
	cf := ClientFunc(func(path string) ([]Requirement, error) {
		if path != "file" {
			t.Fatalf("unexpected path %q", path)
		}
		return []Requirement{{ID: 1, Name: "R"}}, nil
	})
	reqs, err := cf.AnalyzeAttachment("file")
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].ID != 1 || reqs[0].Name != "R" {
		t.Fatalf("unexpected requirements: %#v", reqs)
	}
}
