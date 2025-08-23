# Gemini LLM Client

This package provides a minimal client for Google’s Gemini REST API.  
Its primary goal is to extract potential software requirements from uploaded documents and to answer free-form prompts.

## Key Types & Functions

| Symbol | Purpose |
| --- | --- |
| `type Requirement` | Holds an extracted requirement (`ID`, `Name`, `Description`). Custom `UnmarshalJSON` accepts both numeric and string IDs. |
| `type Client interface` | Abstracts Gemini interactions. Implementations must provide:<br>`AnalyzeAttachment(path string) ([]Requirement, error)`<br>`Ask(prompt string) (string, error)` |
| `type ClientFunc` | Adapter that lets you plug ordinary functions into the `Client` interface (useful for tests/mocks). |
| `func SetClient(c Client) Client` | Swaps the global client implementation. Returns the previous client. |
| `func AnalyzeAttachment(path string) ([]Requirement, error)` | Convenience wrapper calling `client.AnalyzeAttachment`. |
| `func Ask(prompt string) (string, error)` | Convenience wrapper calling `client.Ask`. |
| `type RESTClient` | Default client implementation using Gemini’s REST API. Important methods: `init`, `AnalyzeAttachment`, `Ask`, `upload`, `generateFile`, `generateText`, `generate`. |

## Writing Your Own Backend

To swap out Gemini or integrate with another service, implement the `Client` interface:

```go
package mybackend

import "github.com/rjboer/PMFS/pmfs/llm/gemini"

type LocalLLM struct{}

func (LocalLLM) AnalyzeAttachment(path string) ([]gemini.Requirement, error) {
    // read file, call your model, return requirements
}

func (LocalLLM) Ask(prompt string) (string, error) {
    // forward prompt to your model
}
```

Use your backend:

```go
prev := gemini.SetClient(LocalLLM{})
defer gemini.SetClient(prev)

reqs, err := gemini.AnalyzeAttachment("spec.docx")
answer, err := gemini.Ask("Summarize the spec")
```

Because all consumers call package-level functions, you can transparently swap the client at runtime (or in tests) without touching callers.

## Architecture

- **`Client` interface** defines the contract.
- **`RESTClient`** is the default implementation, handling HTTP requests to Gemini.
- **`SetClient`** allows dependency injection, enabling alternate backends or mocks.
- **`AnalyzeAttachment`** decides whether a file is text or binary, then either:
  - embeds the text directly into a prompt (`generateText`), or
  - uploads the file (`upload`) and references it (`generateFile`).
- **`Ask`** builds a single-prompt request for conversational interactions.
- Both high-level methods ultimately call `generate`, which posts to Gemini’s `generateContent` endpoint and parses the response.

```mermaid
flowchart TD
    A[AnalyzeAttachment(path)] --> B{Is text file?}
    B -- Yes --> C[generateText]
    B -- No --> D[upload file]
    D --> E[generateFile]
    C --> F[Parse JSON to []Requirement]
    E --> F
    F --> G[Return requirements]

    H[Ask(prompt)] --> I[generate]
    I --> J[Return string response]
```

---

**Environment variable**: `GEMINI_API_KEY` must be set for `RESTClient` to work.

