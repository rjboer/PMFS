package gemini

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Requirement represents a potential requirement returned by Gemini.
type Requirement struct {
	ID          int    `json:"id,string"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *Requirement) UnmarshalJSON(data []byte) error {
	type Alias Requirement
	aux := struct {
		ID any `json:"id"`
		*Alias
	}{Alias: (*Alias)(r)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	switch v := aux.ID.(type) {
	case float64:
		r.ID = int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		r.ID = i
	default:
		return fmt.Errorf("invalid id type %T", aux.ID)
	}
	return nil
}

// Client defines the behavior needed to analyze attachments and answer prompts.
type Client interface {
	AnalyzeAttachment(path string) ([]Requirement, error)
	Ask(prompt string) (string, error)
}

// ClientFunc allows using ordinary functions as Client.
// Populate the desired function fields to adapt in tests.
type ClientFunc struct {
	AnalyzeAttachmentFunc func(string) ([]Requirement, error)
	AskFunc               func(string) (string, error)
}

// AnalyzeAttachment satisfies Client interface.
func (f ClientFunc) AnalyzeAttachment(path string) ([]Requirement, error) {
	if f.AnalyzeAttachmentFunc == nil {
		return nil, errors.New("AnalyzeAttachment not implemented")
	}
	return f.AnalyzeAttachmentFunc(path)
}

// Ask satisfies Client interface.
func (f ClientFunc) Ask(prompt string) (string, error) {
	if f.AskFunc == nil {
		return "", errors.New("Ask not implemented")
	}
	return f.AskFunc(prompt)
}

// DefaultClient is the package's default Gemini client.
var (
	DefaultClient Client = &RESTClient{}
	client        Client = DefaultClient
)

// SetClient replaces the package's client, returning the previous one.
func SetClient(c Client) Client {
	old := client
	client = c
	DefaultClient = c
	return old
}

// AnalyzeAttachment uploads and analyzes the file at path using the configured client.
func AnalyzeAttachment(path string) ([]Requirement, error) {
	return client.AnalyzeAttachment(path)
}

// Ask sends a prompt to the configured client and returns the response.
func Ask(prompt string) (string, error) {
	return client.Ask(prompt)
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
	mt := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if i := strings.Index(mt, ";"); i >= 0 {
		mt = mt[:i]
	}
	if strings.HasPrefix(mt, "text/") {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return c.generateText(string(b))
	}

	fileID, mimeType, err := c.upload(path)
	if err != nil {
		return nil, err
	}

	return c.generateFile(fileID, mimeType)
}

// Ask sends a prompt to Gemini and returns the raw text response.
func (c *RESTClient) Ask(prompt string) (string, error) {
	if err := c.init(); err != nil {
		return "", err
	}
	body := map[string]any{
		"contents": []any{map[string]any{
			"parts": []any{map[string]any{"text": prompt}},
		}},
	}
	return c.generate(body)
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

func (c *RESTClient) generateFile(fileID, mimeType string) ([]Requirement, error) {
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

	text, err := c.generate(body)
	if err != nil {
		return nil, err
	}
	var reqs []Requirement
	if err := json.Unmarshal([]byte(text), &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}

func (c *RESTClient) generateText(text string) ([]Requirement, error) {
	prompt := `You are an assistant that extracts potential software requirements from files.
Return a JSON array of objects with fields "id", "name", and "description".`

	body := map[string]any{
		"contents": []any{map[string]any{
			"parts": []any{
				map[string]any{"text": text},
				map[string]any{"text": prompt},
			},
		}},
		"generationConfig": map[string]any{"responseMimeType": "application/json"},
	}

	resp, err := c.generate(body)
	if err != nil {
		return nil, err
	}
	var reqs []Requirement
	if err := json.Unmarshal([]byte(resp), &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}

func (c *RESTClient) generate(body map[string]any) (string, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=%s", c.APIKey)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("generate failed: %s: %s", resp.Status, string(rb))
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
		return "", err
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from gemini")
	}
	return gr.Candidates[0].Content.Parts[0].Text, nil
}
