package gemini

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Requirement represents a potential requirement returned by Gemini.
type Requirement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Client defines the behavior needed to analyze attachments.
type Client interface {
	AnalyzeAttachment(path string) ([]Requirement, error)
}

// ClientFunc allows using ordinary functions as Client.
type ClientFunc func(string) ([]Requirement, error)

// AnalyzeAttachment satisfies Client interface.
func (f ClientFunc) AnalyzeAttachment(path string) ([]Requirement, error) {
	return f(path)
}

var client Client = &RESTClient{}

// SetClient replaces the package's client, returning the previous one.
func SetClient(c Client) Client {
	old := client
	client = c
	return old
}

// AnalyzeAttachment uploads and analyzes the file at path using the configured client.
func AnalyzeAttachment(path string) ([]Requirement, error) {
	return client.AnalyzeAttachment(path)
}

// RESTClient implements Client using Gemini's REST API.
type RESTClient struct {
	HTTPClient *http.Client
	APIKey     string
}

func (c *RESTClient) init() error {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	if c.APIKey == "" {
		c.APIKey = os.Getenv("GEMINI_API_KEY")
		if c.APIKey == "" {
			return errors.New("GEMINI_API_KEY not set")
		}
	}
	return nil
}

// AnalyzeAttachment implements the upload and generation flow.
func (c *RESTClient) AnalyzeAttachment(path string) ([]Requirement, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	fileID, mimeType, err := c.upload(path)
	if err != nil {
		return nil, err
	}

	return c.generate(fileID, mimeType)
}

func (c *RESTClient) upload(path string) (fileID, mimeType string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return "", "", err
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", "", err
	}
	if err := w.Close(); err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/upload/v1beta/files?key=%s", c.APIKey)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("upload failed: %s: %s", resp.Status, string(b))
	}

	var ur struct {
		File struct {
			Name     string `json:"name"`
			MimeType string `json:"mimeType"`
		} `json:"file"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ur); err != nil {
		return "", "", err
	}
	return ur.File.Name, ur.File.MimeType, nil
}

func (c *RESTClient) generate(fileID, mimeType string) ([]Requirement, error) {
	prompt := `You are an assistant that extracts potential software requirements from files.
Return a JSON array of objects with fields "id", "name", and "description".`

	body := map[string]any{
		"contents": []any{map[string]any{
			"parts": []any{
				map[string]any{
					"file_data": map[string]any{
						"file_uri":  fileID,
						"mime_type": mimeType,
					},
				},
				map[string]any{"text": prompt},
			},
		}},
		"generationConfig": map[string]any{"responseMimeType": "application/json"},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=%s", c.APIKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate failed: %s: %s", resp.Status, string(rb))
	}

	var gr struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil, err
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response from gemini")
	}

	var reqs []Requirement
	if err := json.Unmarshal([]byte(gr.Candidates[0].Content.Parts[0].Text), &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}
