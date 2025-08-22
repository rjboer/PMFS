package config

import (
	"bufio"
	"os"
	"strings"
)

// init loads environment variables from a local .env file if present.
// Existing variables are not overridden.
func init() {
	loadDotEnv()
}

// loadDotEnv reads .env from the current working directory and sets
// environment variables defined as KEY=VALUE pairs. Lines beginning with '#'
// or empty lines are ignored. Existing environment variables take precedence
// over values in the file.
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return // silently ignore missing .env
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
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}
