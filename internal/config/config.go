package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config holds the CLI configuration.
type Config struct {
	// Env selects a preset in EnvPresets (dev|test|prod). Ignored if APIURL / ZHIZAI_API_URL is set.
	Env string `json:"env,omitempty"`
	// APIURL is an absolute business API base URL override.
	APIURL string `json:"api_url,omitempty"`
	// OAuthURL is an absolute OAuth2 API base URL override (authorize/token/refresh).
	OAuthURL string `json:"oauth_url,omitempty"`

	AuthMode string `json:"auth_mode,omitempty"` // api_key | oauth

	APIKey    string `json:"api_key,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"` // legacy / display for api key
	TeamID    string `json:"team_id,omitempty"`

	AccessToken    string `json:"access_token,omitempty"`
	RefreshToken   string `json:"refresh_token,omitempty"`
	TokenType      string `json:"token_type,omitempty"`
	TokenExpiresAt int64  `json:"token_expires_at,omitempty"` // unix seconds
	Scope          string `json:"scope,omitempty"`
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

	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	c.normalizeAuthMode()
	return nil
}

func (c *Config) normalizeAuthMode() {
	if c.AuthMode != "" {
		return
	}
	if c.AccessToken != "" || c.RefreshToken != "" {
		c.AuthMode = AuthModeOAuth
		return
	}
	if c.APIKey != "" {
		c.AuthMode = AuthModeAPIKey
	}
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

// Clear removes stored credentials and saves the config (keeps env / api_url).
func (c *Config) Clear() error {
	c.AuthMode = ""
	c.APIKey = ""
	c.ExpiresAt = ""
	c.TeamID = ""
	c.AccessToken = ""
	c.RefreshToken = ""
	c.TokenType = ""
	c.TokenExpiresAt = 0
	c.Scope = ""
	return c.Save()
}

// SetAPIKeyLogin stores API Key mode and clears OAuth tokens (mutual exclusion).
func (c *Config) SetAPIKeyLogin(apiKey string) error {
	c.AuthMode = AuthModeAPIKey
	c.APIKey = apiKey
	c.AccessToken = ""
	c.RefreshToken = ""
	c.TokenType = ""
	c.TokenExpiresAt = 0
	c.Scope = ""
	return c.Save()
}

// SetOAuthLogin stores OAuth tokens and clears API Key (mutual exclusion).
func (c *Config) SetOAuthLogin(accessToken, refreshToken, tokenType, scope string, expiresIn int) error {
	c.AuthMode = AuthModeOAuth
	c.APIKey = ""
	c.ExpiresAt = ""
	c.AccessToken = accessToken
	c.RefreshToken = refreshToken
	c.TokenType = tokenType
	c.Scope = scope
	if expiresIn > 0 {
		c.TokenExpiresAt = time.Now().Unix() + int64(expiresIn)
	} else {
		c.TokenExpiresAt = 0
	}
	return c.Save()
}

// UpdateOAuthTokens updates tokens after refresh (rotation-safe).
func (c *Config) UpdateOAuthTokens(accessToken, refreshToken, tokenType, scope string, expiresIn int) error {
	c.AuthMode = AuthModeOAuth
	c.AccessToken = accessToken
	if refreshToken != "" {
		c.RefreshToken = refreshToken
	}
	if tokenType != "" {
		c.TokenType = tokenType
	}
	if scope != "" {
		c.Scope = scope
	}
	if expiresIn > 0 {
		c.TokenExpiresAt = time.Now().Unix() + int64(expiresIn)
	}
	return c.Save()
}

// IsLoggedIn reports whether any usable credential is configured.
func (c *Config) IsLoggedIn() bool {
	c.normalizeAuthMode()
	switch c.AuthMode {
	case AuthModeOAuth:
		return c.AccessToken != "" || c.RefreshToken != ""
	case AuthModeAPIKey:
		return c.APIKey != ""
	default:
		return c.APIKey != "" || c.AccessToken != "" || c.RefreshToken != ""
	}
}

// TokenExpired reports whether the access token is past TokenExpiresAt (with skew).
func (c *Config) TokenExpired(skewSeconds int64) bool {
	if c.TokenExpiresAt <= 0 {
		return false
	}
	return time.Now().Unix() >= c.TokenExpiresAt-skewSeconds
}

// ResetForTests clears the singleton so tests can start fresh.
func ResetForTests() {
	instance = nil
	once = sync.Once{}
}
