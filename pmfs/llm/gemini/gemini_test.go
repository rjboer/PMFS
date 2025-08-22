package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestRESTClientAnalyzeAttachmentSpecs(t *testing.T) {
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
	c := &RESTClient{HTTPClient: &http.Client{Transport: mockTransport{}}}
	specs := []string{"spec1.txt", "spec2.txt", "spec3.png", "spec4.jpg"}
	base := filepath.Join("..", "..", "..", "testdata")
	for _, s := range specs {
		s := s
		t.Run(s, func(t *testing.T) {
			p := filepath.Join(base, s)
			reqs, err := c.AnalyzeAttachment(p)
			if err != nil {
				t.Fatalf("AnalyzeAttachment(%s): %v", s, err)
			}
			t.Logf("Gemini returned for %s: %#v", s, reqs)
			if len(reqs) != 1 || reqs[0].Name != "R1" {
				t.Fatalf("unexpected requirements: %#v", reqs)
			}
		})
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

type mockGeminiTransport struct {
	mu     sync.Mutex
	nextID int
	files  map[string]string
}

func newMockGeminiTransport() *mockGeminiTransport {
	return &mockGeminiTransport{files: make(map[string]string), nextID: 1}
}

func (m *mockGeminiTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "/upload/") {
		boundary := strings.TrimPrefix(req.Header.Get("Content-Type"), "multipart/form-data; boundary=")
		mr := multipart.NewReader(req.Body, boundary)
		var fname string
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if part.FormName() == "file" {
				fname = part.FileName()
				break
			}
		}
		m.mu.Lock()
		id := fmt.Sprintf("files/%d", m.nextID)
		m.nextID++
		var resp string
		switch {
		case strings.Contains(fname, "spec1"):
			resp = `[{"id":1,"name":"Spec1"}]`
		case strings.Contains(fname, "spec2"):
			resp = `[{"id":1,"name":"Spec2"}]`
		default:
			resp = ""
		}
		m.files[id] = resp
		m.mu.Unlock()

		rec := httptest.NewRecorder()
		rec.WriteString(fmt.Sprintf(`{"file":{"name":"%s","mimeType":"text/plain"}}`, id))
		return rec.Result(), nil
	}
	if strings.Contains(req.URL.Path, ":generateContent") {
		var body struct {
			Contents []struct {
				Parts []struct {
					FileData struct {
						FileURI string `json:"file_uri"`
					} `json:"file_data"`
				} `json:"parts"`
			} `json:"contents"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			return nil, err
		}
		id := body.Contents[0].Parts[0].FileData.FileURI
		m.mu.Lock()
		resp, ok := m.files[id]
		m.mu.Unlock()
		rec := httptest.NewRecorder()
		if !ok {
			rec.Code = http.StatusBadRequest
			rec.Body = bytes.NewBufferString("unknown file")
			return rec.Result(), nil
		}
		if resp == "" {
			rec.Code = http.StatusInternalServerError
			rec.Body = bytes.NewBufferString("bad file")
			return rec.Result(), nil
		}
		rec.WriteString(fmt.Sprintf(`{"candidates":[{"content":{"parts":[{"text":%q}]}}]}`, resp))
		return rec.Result(), nil
	}
	return nil, fmt.Errorf("unexpected path %s", req.URL.Path)
}

func sameRequirements(a, b []Requirement) bool {
	if len(a) != len(b) {
		return false
	}
	ba, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return bytes.Equal(ba, bb)
}

func TestRESTClientAnalyzeAttachmentReal(t *testing.T) {
	base := filepath.Join("..", "..", "..", "testdata")
	p1 := filepath.Join(base, "spec1.txt")
	p2 := filepath.Join(base, "spec2.txt")

	key := os.Getenv("GEMINI_API_KEY")
	if key != "" && key != "test-key" {
		c := &RESTClient{}
		r1, err := c.AnalyzeAttachment(p1)
		if err != nil {
			t.Fatalf("AnalyzeAttachment(spec1): %v", err)
		}
		t.Logf("real Gemini returned for spec1: %#v", r1)
		r2, err := c.AnalyzeAttachment(p2)
		if err != nil {
			t.Fatalf("AnalyzeAttachment(spec2): %v", err)
		}
		t.Logf("real Gemini returned for spec2: %#v", r2)
		if sameRequirements(r1, r2) {
			t.Fatalf("expected different requirements for distinct documents")
		}
		return
	}

	t.Setenv("GEMINI_API_KEY", "")
	mt := newMockGeminiTransport()
	c := &RESTClient{APIKey: "k", HTTPClient: &http.Client{Transport: mt}}
	r1, err := c.AnalyzeAttachment(p1)
	if err != nil {
		t.Fatalf("AnalyzeAttachment(spec1): %v", err)
	}
	t.Logf("mock Gemini returned for spec1: %#v", r1)
	r2, err := c.AnalyzeAttachment(p2)
	if err != nil {
		t.Fatalf("AnalyzeAttachment(spec2): %v", err)
	}
	t.Logf("mock Gemini returned for spec2: %#v", r2)
	if sameRequirements(r1, r2) {
		t.Fatalf("expected different requirements for mock documents")
	}

	if _, err := c.AnalyzeAttachment(filepath.Join(base, "sources.txt")); err == nil {
		t.Fatalf("expected error for unsupported document")
	}
}
