package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BoteAI/zhizai-cli/internal/config"
)

const (
	minRequestInterval = 500 * time.Millisecond
	maxNetworkRetries  = 3
)

// APIError is the CLI-facing error shape for -o json output.
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Reason    string `json:"reason"`
	Retryable bool   `json:"retryable"`
}

// RequestError is returned when the API call fails.
type RequestError struct {
	APIError
	StatusCode int
}

func (e *RequestError) Error() string {
	parts := []string{fmt.Sprintf("API error %s", e.Code)}
	if e.Reason != "" {
		parts = append(parts, e.Reason)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	return strings.Join(parts, ": ")
}

// Client is an HTTP client for the 智在记录 OpenAPI.
type Client struct {
	baseURL        string
	oauthURL       string
	authMode       string
	apiKey         string
	accessToken    string
	refreshToken   string
	tokenExpiresAt int64
	httpClient     *http.Client
	limiter        sync.Mutex
	lastReq        time.Time
	refreshedOnce  bool
}

// New creates a new API client from config / environment.
func New() *Client {
	cfg := config.Get()
	baseURL := config.ResolveAPIBaseURL(cfg)
	c := NewWithOptions(baseURL, "", nil)
	c.oauthURL = config.ResolveOAuthBaseURL(cfg)
	c.applyFromConfig(cfg)
	return c
}

func (c *Client) applyFromConfig(cfg *config.Config) {
	envKey := strings.TrimSpace(os.Getenv("ZHIZAI_REC_API_KEY"))
	if envKey != "" {
		c.authMode = config.AuthModeAPIKey
		c.apiKey = envKey
		c.accessToken = ""
		c.refreshToken = ""
		c.tokenExpiresAt = 0
		return
	}

	switch cfg.AuthMode {
	case config.AuthModeOAuth:
		c.authMode = config.AuthModeOAuth
		c.accessToken = cfg.AccessToken
		c.refreshToken = cfg.RefreshToken
		c.tokenExpiresAt = cfg.TokenExpiresAt
		c.apiKey = ""
	default:
		c.authMode = config.AuthModeAPIKey
		c.apiKey = cfg.APIKey
		c.accessToken = ""
		c.refreshToken = ""
		c.tokenExpiresAt = 0
	}
}

// NewWithOptions creates a client with explicit API key (api_key mode). Useful for tests.
func NewWithOptions(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	mode := ""
	if apiKey != "" {
		mode = config.AuthModeAPIKey
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		oauthURL:   strings.TrimRight(baseURL, "/"), // tests default OAuth to same host
		authMode:   mode,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// NewOAuthWithOptions creates an OAuth-mode client for tests.
func NewOAuthWithOptions(baseURL, accessToken, refreshToken string, httpClient *http.Client) *Client {
	c := NewWithOptions(baseURL, "", httpClient)
	c.authMode = config.AuthModeOAuth
	c.accessToken = accessToken
	c.refreshToken = refreshToken
	return c
}

// SetOAuthBaseURL overrides the OAuth API root (authorize/token/refresh).
func (c *Client) SetOAuthBaseURL(oauthURL string) {
	c.oauthURL = strings.TrimRight(strings.TrimSpace(oauthURL), "/")
}

func (c *Client) resolveAPIURL(path string) string {
	return config.JoinAPIURL(c.baseURL, path)
}

// APIEndpoint returns the absolute URL for a relative business API path.
func (c *Client) APIEndpoint(path string) string {
	return c.resolveAPIURL(path)
}

func (c *Client) resolveOAuthURL(path string) string {
	base := c.oauthURL
	if base == "" {
		base = c.baseURL
	}
	return config.JoinOAuthURL(base, path)
}

func (c *Client) waitRateLimit() {
	c.limiter.Lock()
	defer c.limiter.Unlock()
	if !c.lastReq.IsZero() {
		wait := minRequestInterval - time.Since(c.lastReq)
		if wait > 0 {
			time.Sleep(wait)
		}
	}
	c.lastReq = time.Now()
}

type apiEnvelope struct {
	ResultCode   string          `json:"resultCode"`
	ResultMsg    string          `json:"resultMsg"`
	ResultObject json.RawMessage `json:"resultObject"`
	Raw          string          `json:"-"`
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	if err := c.EnsureAccessToken(); err != nil {
		return nil, err
	}

	if c.authMode == config.AuthModeOAuth {
		if c.accessToken == "" {
			return nil, &RequestError{
				APIError: APIError{
					Code:      "unauthorized",
					Message:   "未配置 OAuth access_token，请运行 zhizai auth login",
					Reason:    "missing_access_token",
					Retryable: false,
				},
			}
		}
	} else if c.apiKey == "" {
		return nil, &RequestError{
			APIError: APIError{
				Code:      "unauthorized",
				Message:   "未配置凭证，请运行 zhizai auth login",
				Reason:    "missing_credentials",
				Retryable: false,
			},
		}
	}

	reqURL := c.resolveAPIURL(path)
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}
	if c.authMode == config.AuthModeOAuth {
		req.Header.Set("X-OAuth2-Access-Token", "Bearer "+c.accessToken)
	} else {
		req.Header.Set("Authorization", c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, token := range []string{
		"connection reset by peer",
		"connection refused",
		"broken pipe",
		"i/o timeout",
		"tls handshake timeout",
		"server closed idle connection",
		"temporary failure",
		"network is unreachable",
	} {
		if strings.Contains(msg, token) {
			return true
		}
	}
	return errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE)
}

func networkRequestError(err error) error {
	return &RequestError{
		APIError: APIError{
			Code:      "network_error",
			Message:   err.Error(),
			Reason:    "connection_failed",
			Retryable: true,
		},
	}
}

func (c *Client) do(method, path string, body []byte) (json.RawMessage, error) {
	var lastErr error
	for attempt := 0; attempt <= maxNetworkRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
		}
		c.waitRateLimit()

		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := c.newRequest(method, path, reader)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if isRetryableNetworkError(err) && attempt < maxNetworkRetries {
				continue
			}
			return nil, networkRequestError(err)
		}

		raw, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			if isRetryableNetworkError(err) && attempt < maxNetworkRetries {
				continue
			}
			return nil, networkRequestError(err)
		}

		if (resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 406) &&
			c.authMode == config.AuthModeOAuth && c.refreshToken != "" && !c.refreshedOnce {
			c.refreshedOnce = true
			if refreshErr := c.forceRefresh(); refreshErr == nil {
				attempt--
				continue
			}
		}

		if resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 406 {
			return nil, &RequestError{
				APIError: APIError{
					Code:      fmt.Sprintf("%d", resp.StatusCode),
					Message:   "无权限或请求无效，请检查登录状态（zhizai auth login / ZHIZAI_REC_API_KEY）",
					Reason:    "unauthorized",
					Retryable: false,
				},
				StatusCode: resp.StatusCode,
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			retryable := resp.StatusCode >= 500 || resp.StatusCode == 429
			if retryable && attempt < maxNetworkRetries {
				lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				continue
			}
			return nil, &RequestError{
				APIError: APIError{
					Code:      fmt.Sprintf("%d", resp.StatusCode),
					Message:   string(raw),
					Reason:    "http_error",
					Retryable: retryable,
				},
				StatusCode: resp.StatusCode,
			}
		}

		var envelope apiEnvelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return nil, fmt.Errorf("parsing API response: %w", err)
		}

		if envelope.ResultCode != "0" {
			retryable := envelope.ResultCode == "429"
			if retryable && attempt < maxNetworkRetries {
				lastErr = fmt.Errorf("%s", envelope.ResultMsg)
				continue
			}
			return nil, &RequestError{
				APIError: APIError{
					Code:      envelope.ResultCode,
					Message:   envelope.ResultMsg,
					Reason:    "api_error",
					Retryable: retryable,
				},
				StatusCode: resp.StatusCode,
			}
		}

		if len(envelope.ResultObject) == 0 {
			return json.RawMessage("null"), nil
		}
		return envelope.ResultObject, nil
	}
	if lastErr != nil {
		return nil, networkRequestError(lastErr)
	}
	return nil, fmt.Errorf("request failed after retries")
}

func (c *Client) forceRefresh() error {
	token, err := c.RefreshAccessToken(c.refreshToken)
	if err != nil {
		return err
	}
	c.accessToken = token.AccessToken
	if token.RefreshToken != "" {
		c.refreshToken = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		c.tokenExpiresAt = time.Now().Unix() + int64(token.ExpiresIn)
	}
	return config.Get().UpdateOAuthTokens(token.AccessToken, token.RefreshToken, token.TokenType, token.Scope, int(token.ExpiresIn))
}

func doGet(c *Client, path string) (json.RawMessage, error) {
	return c.do(http.MethodGet, path, nil)
}

func doPost(c *Client, path string, payload any) (json.RawMessage, error) {
	var body []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = data
	}
	return c.do(http.MethodPost, path, body)
}

// Ping verifies credentials with a minimal note list query.
func (c *Client) Ping() error {
	_, err := c.NoteList(NoteListParams{PageNum: 1, PageSize: 1})
	return err
}
