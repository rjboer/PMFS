package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRESTClientInitUsesAPIKeyFromEnv(t *testing.T) {
	//check if we can use the .env
	if b, err := os.ReadFile(filepath.Join("..", "..", "..", ".env")); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "GEMINI_API_KEY=") {
				key := strings.TrimSpace(strings.TrimPrefix(line, "GEMINI_API_KEY="))
				os.Setenv("GEMINI_API_KEY", key)
				break
			}
		}
	}

	key, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok || key == "" {
		t.Skip("GEMINI_API_KEY not set")
	}
	c := &RESTClient{}
	if err := c.init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if c.APIKey != key {
		t.Fatalf("expected APIKey %q, got %q", key, c.APIKey)
	}
}

type testClient struct{ name string }

func (t testClient) AnalyzeAttachment(path string) ([]Requirement, error) {
	return []Requirement{{Name: t.name}}, nil
}

func (t testClient) Ask(prompt string) (string, error) {
	return t.name, nil
}

func TestClientFuncAnalyzeAttachment(t *testing.T) {
	cf := ClientFunc{AnalyzeAttachmentFunc: func(path string) ([]Requirement, error) {
		if path != "file" {
			t.Fatalf("unexpected path %q", path)
		}
		return []Requirement{{ID: 1, Name: "R"}}, nil
	}}
	reqs, err := cf.AnalyzeAttachment("file")
	if err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
	if len(reqs) != 1 || reqs[0].ID != 1 || reqs[0].Name != "R" {
		t.Fatalf("unexpected requirements: %#v", reqs)
	}
}

func TestClientFuncAsk(t *testing.T) {
	cf := ClientFunc{AskFunc: func(prompt string) (string, error) {
		if prompt != "p" {
			t.Fatalf("unexpected prompt %q", prompt)
		}
		return "answer", nil
	}}
	ans, err := cf.Ask("p")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if ans != "answer" {
		t.Fatalf("unexpected answer %q", ans)
	}
}

func TestRequirementUnmarshalStringID(t *testing.T) {
	data := []byte(`[{"id":"42","name":"N","description":"D"}]`)
	var reqs []Requirement
	if err := json.Unmarshal(data, &reqs); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(reqs) != 1 || reqs[0].ID != 42 {
		t.Fatalf("unexpected requirements: %#v", reqs)
	}
}

func TestRequirementUnmarshalNonNumericID(t *testing.T) {
	data := []byte(`[{"id":"REQ-1","name":"N","description":"D"}]`)
	var reqs []Requirement
	if err := json.Unmarshal(data, &reqs); err == nil {
		t.Fatalf("expected error for non-numeric id")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAnalyzeAttachmentUsesUploadURI(t *testing.T) {
	expectedURI := "https://generativelanguage.googleapis.com/v1beta/files/abc123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/upload/v1beta/files":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, fmt.Sprintf(`{"file":{"name":"files/abc123","mimeType":"image/png","uri":%q}}`, expectedURI))
		case "/v1beta/models/gemini-1.5-flash-latest:generateContent":
			var body struct {
				Contents []struct {
					Parts []struct {
						FileData struct {
							FileURI string `json:"file_uri"`
						} `json:"file_data"`
					} `json:"parts"`
				} `json:"contents"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			got := body.Contents[0].Parts[0].FileData.FileURI
			if got != expectedURI {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"[]"}]}}]}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	c := &RESTClient{
		APIKey: "test-key",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			return http.DefaultTransport.RoundTrip(req)
		})},
	}

	tmp, err := os.CreateTemp(t.TempDir(), "file*.png")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := tmp.Write([]byte("data")); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	tmp.Close()

	if _, err := c.AnalyzeAttachment(tmp.Name()); err != nil {
		t.Fatalf("AnalyzeAttachment: %v", err)
	}
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
	p3 := filepath.Join(base, "spec3.png")
	p4 := filepath.Join(base, "spec4.jpg")

	key := os.Getenv("GEMINI_API_KEY")
	if key == "" || key == "test-key" {
		t.Skip("GEMINI_API_KEY not set")
	}

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

	r3, err := c.AnalyzeAttachment(p3)
	if err != nil {
		t.Fatalf("AnalyzeAttachment(spec3): %v", err)
	}
	t.Logf("real Gemini returned for spec3: %#v", r3)

	r4, err := c.AnalyzeAttachment(p4)
	if err != nil {
		t.Fatalf("AnalyzeAttachment(spec4): %v", err)
	}
	t.Logf("real Gemini returned for spec4: %#v", r4)

	if sameRequirements(r1, r2) {
		t.Fatalf("expected different requirements for distinct documents")
	}

	if _, err := c.AnalyzeAttachment(filepath.Join(base, "sources.txt")); err == nil {
		t.Fatalf("expected error for unsupported document")
	}
}
