package llm

import (
	"time"

	"github.com/rjboer/PMFS/pmfs/llm/gemini"
)

type rateLimitedClient struct {
	Client
	tick <-chan time.Time
}

func NewRateLimitedClient(c Client, rps int) Client {
	if rps <= 0 {
		rps = 1
	}
	interval := time.Second / time.Duration(rps)
	return &rateLimitedClient{Client: c, tick: time.Tick(interval)}
}

func (r *rateLimitedClient) wait() {
	<-r.tick
}

func (r *rateLimitedClient) Ask(prompt string) (string, error) {
	r.wait()
	return r.Client.Ask(prompt)
}

func (r *rateLimitedClient) AnalyzeAttachment(path string) ([]gemini.Requirement, error) {
	r.wait()
	return r.Client.AnalyzeAttachment(path)
}
