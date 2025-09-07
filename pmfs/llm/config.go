package llm

import (
	"encoding/json"
	"os"

	"github.com/rjboer/PMFS/pmfs/llm/gemini"
)

type Config struct {
	Model             string `json:"model"`
	RequestsPerSecond int    `json:"requests_per_second"`
}

func LoadConfig() Config {
	cfg := Config{
		Model:             gemini.DefaultModel,
		RequestsPerSecond: 3,
	}
	if b, err := os.ReadFile("llmconfig.json"); err == nil {
		_ = json.Unmarshal(b, &cfg)
	}
	if cfg.Model == "" {
		cfg.Model = gemini.DefaultModel
	}
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 3
	}
	return cfg
}

var config = LoadConfig()

func Model() string { return config.Model }

func RequestsPerSecond() int { return config.RequestsPerSecond }
