package gemini

import "errors"

// QAClient defines the ability to ask a question.
type QAClient interface {
	Ask(role, prompt string) (string, error)
}

// QAClientFunc adapts ordinary functions to QAClient.
type QAClientFunc func(role, prompt string) (string, error)

// Ask satisfies the QAClient interface.
func (f QAClientFunc) Ask(role, prompt string) (string, error) {
	return f(role, prompt)
}

// Client is the package level QAClient used for interactive questions.
var Client QAClient = QAClientFunc(func(role, prompt string) (string, error) {
	return "", errors.New("gemini QA client not set")
})

// SetQAClient replaces the package's QA client, returning the previous one.
func SetQAClient(c QAClient) QAClient {
	old := Client
	Client = c
	return old
}
