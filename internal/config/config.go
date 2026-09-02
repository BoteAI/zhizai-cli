package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const DefaultAPIBaseURL = "https://openapi.zzjilu.com/api/v1"

// Config holds the CLI configuration.
type Config struct {
	APIKey    string `json:"api_key"`
	ExpiresAt string `json:"expires_at,omitempty"`
	TeamID    string `json:"team_id,omitempty"`
}

var (
	instance *Config
	once     sync.Once
)

// Get returns the singleton config, loading from file on first call.
func Get() *Config {
	once.Do(func() {
		instance = &Config{}
		_ = instance.load()
	})
	return instance
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zhizai", "config.json"), nil
}

func (c *Config) load() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

// Save writes the current config to disk.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// Clear removes the stored API key and saves the config.
func (c *Config) Clear() error {
	c.APIKey = ""
	c.ExpiresAt = ""
	c.TeamID = ""
	return c.Save()
}

// IsLoggedIn reports whether an API key is configured.
func (c *Config) IsLoggedIn() bool {
	return c.APIKey != ""
}

// ResetForTests clears the singleton so tests can start fresh.
func ResetForTests() {
	instance = nil
	once = sync.Once{}
}
