// Package config resolves prwlrctl settings from flags, environment
// variables, and an optional config file, in that priority order.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	EnvBaseURL  = "PROWLER_BASE_URL"
	EnvAPIKey   = "PROWLER_API_KEY"
	EnvToken    = "PROWLER_TOKEN" // JWT access token, takes precedence over API key
	DefaultBase = "https://api.prowler.com/api/v1"
)

type Config struct {
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Path returns ~/.config/prwlrctl/config.json (respects $XDG_CONFIG_HOME).
func Path() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "prwlrctl", "config.json"), nil
}

// Load reads the config file if present; missing file is not an error.
func Load() (Config, error) {
	var cfg Config
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes the config file with restrictive permissions (it may hold
// secrets: API key / JWT tokens).
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Resolve merges file config with environment variables and explicit flag
// overrides (flags win, then env vars, then file, then default).
func Resolve(flagBaseURL, flagAPIKey, flagToken string) Config {
	cfg, _ := Load()

	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv(EnvToken); v != "" {
		cfg.AccessToken = v
	}

	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}
	if flagToken != "" {
		cfg.AccessToken = flagToken
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBase
	}
	return cfg
}
