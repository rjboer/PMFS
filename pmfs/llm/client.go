package llm

import (
	"os"

	"github.com/rjboer/PMFS/pmfs/llm/gemini"
)

// Client defines the behavior needed to analyze attachments and answer prompts.
type Client interface {
	AnalyzeAttachment(path string) ([]gemini.Requirement, error)
	Ask(prompt string) (string, error)
}

var (
	// DefaultClient is the package's default LLM client wrapped with a rate limiter.
	DefaultClient Client = NewRateLimitedClient(
		gemini.NewRESTClient(os.Getenv("GEMINI_API_KEY"), config.Model),
		config.RequestsPerSecond,
	)
	client Client = DefaultClient
)

// SetClient replaces the package's client, returning the previous one.
// This is intended for internal testing use only.
func SetClient(c Client) Client {
	old := client
	client = c
	DefaultClient = c
	return old
}

// AnalyzeAttachment uploads and analyzes the file at path using the configured client.
func AnalyzeAttachment(path string) ([]gemini.Requirement, error) {
	return client.AnalyzeAttachment(path)
}

// Ask sends a prompt to the configured client and returns the response.
func Ask(prompt string) (string, error) {
	return client.Ask(prompt)
}
