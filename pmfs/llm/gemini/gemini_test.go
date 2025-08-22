package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadEnv reads a .env file in dir and sets env vars.
func loadEnv(dir string) error {
	f, err := os.Open(filepath.Join(dir, ".env"))
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		os.Setenv(key, val)
	}
	return scanner.Err()
}

func TestRESTClientInitUsesAPIKeyFromEnvFile(t *testing.T) {
	dir := t.TempDir()
	env := "GEMINI_API_KEY=test-key\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := loadEnv(dir); err != nil {
		t.Fatalf("loadEnv: %v", err)
	}
	c := &RESTClient{}
	if err := c.init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if c.APIKey != "test-key" {
		t.Fatalf("expected APIKey 'test-key', got %q", c.APIKey)
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
