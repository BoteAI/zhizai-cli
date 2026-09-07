package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BoteAI/zhizai-cli/internal/config"
)

const (
	deviceGrantType  = "urn:ietf:params:oauth:grant-type:device_code"
	refreshGrantType = "refresh_token"
)

// flexInt unmarshals JSON numbers or numeric strings.
type flexInt int

func (v *flexInt) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*v = 0
		return nil
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		if strings.TrimSpace(str) == "" {
			*v = 0
			return nil
		}
		n, err := strconv.Atoi(strings.TrimSpace(str))
		if err != nil {
			return fmt.Errorf("flexInt: %w", err)
		}
		*v = flexInt(n)
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*v = flexInt(n)
	return nil
}

// DeviceAuthSession is returned by DeviceAuthorize.
type DeviceAuthSession struct {
	DeviceCode              string  `json:"device_code"`
	UserCode                string  `json:"user_code"`
	VerificationURI         string  `json:"verification_uri"`
	VerificationURIComplete string  `json:"verification_uri_complete"`
	ExpiresIn               flexInt `json:"expires_in"`
	Interval                flexInt `json:"interval"`
}

// TokenResponse is an OAuth token payload.
type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    flexInt `json:"expires_in"`
	RefreshToken string  `json:"refresh_token"`
	Scope        string  `json:"scope"`
}

// DeviceAuthorize starts an OAuth 2.0 device authorization session.
// scope is optional; empty means server default scopes.
// The second return is the raw JSON body for debugging.
func (c *Client) DeviceAuthorize(scope string) (*DeviceAuthSession, string, error) {
	form := url.Values{}
	form.Set("client_id", config.OAuthClientID)
	if strings.TrimSpace(scope) != "" {
		form.Set("scope", strings.TrimSpace(scope))
	}

	env, err := c.postOAuthForm("/oauth2/device/authorize", form)
	if err != nil {
		return nil, "", err
	}
	raw := env.Raw
	if env.ResultCode != "0" {
		return nil, raw, &RequestError{
			APIError: APIError{
				Code:      env.ResultCode,
				Message:   env.ResultMsg,
				Reason:    "device_authorize_failed",
				Retryable: false,
			},
		}
	}

	var session DeviceAuthSession
	if err := json.Unmarshal(env.ResultObject, &session); err != nil {
		return nil, raw, fmt.Errorf("parsing device authorize response: %w", err)
	}
	if session.DeviceCode == "" || session.UserCode == "" {
		return nil, raw, fmt.Errorf("device authorize response missing device_code/user_code")
	}
	if session.Interval <= 0 {
		session.Interval = 5
	}
	if session.ExpiresIn <= 0 {
		session.ExpiresIn = 600
	}
	return &session, raw, nil
}

// PollDeviceToken polls once for a device_code grant.
// Returns (nil, pendingMsg, raw, nil) when still pending.
func (c *Client) PollDeviceToken(deviceCode string) (*TokenResponse, string, string, error) {
	form := url.Values{}
	form.Set("grant_type", deviceGrantType)
	form.Set("client_id", config.OAuthClientID)
	form.Set("device_code", deviceCode)

	env, err := c.postOAuthForm("/oauth2/token", form)
	if err != nil {
		return nil, "", "", err
	}
	raw := env.Raw

	msg := strings.TrimSpace(env.ResultMsg)
	if env.ResultCode == "0" {
		token, err := decodeTokenObject(env.ResultObject)
		if err != nil {
			return nil, "", raw, err
		}
		return token, "", raw, nil
	}

	kind := classifyDevicePollStatus(env.ResultCode, msg)
	switch kind {
	case pollStatusPending, pollStatusSlowDown:
		return nil, kind, raw, nil
	case pollStatusAccessDenied, pollStatusExpiredToken, pollStatusInvalidGrant,
		pollStatusInvalidClient, pollStatusUnauthorizedClient, pollStatusInvalidScope,
		pollStatusUnsupportedGrant, pollStatusInvalidRequest, pollStatusServerError:
		return nil, kind, raw, &RequestError{
			APIError: APIError{
				Code:      env.ResultCode,
				Message:   msg,
				Reason:    kind,
				Retryable: false,
			},
		}
	default:
		return nil, kind, raw, &RequestError{
			APIError: APIError{
				Code:      env.ResultCode,
				Message:   env.ResultMsg,
				Reason:    "token_poll_failed",
				Retryable: false,
			},
		}
	}
}

// Device poll status kinds (stable internal names; branch primarily on resultCode).
const (
	pollStatusPending            = "authorization_pending"
	pollStatusSlowDown           = "slow_down"
	pollStatusAccessDenied       = "access_denied"
	pollStatusExpiredToken       = "expired_token"
	pollStatusInvalidGrant       = "invalid_grant"
	pollStatusInvalidClient      = "invalid_client"
	pollStatusUnauthorizedClient = "unauthorized_client"
	pollStatusInvalidScope       = "invalid_scope"
	pollStatusUnsupportedGrant   = "unsupported_grant_type"
	pollStatusInvalidRequest     = "invalid_request"
	pollStatusServerError        = "server_error"
	pollStatusUnknown            = "unknown"
)

// Platform NoteErrorConstant codes for OAuth2 device token polling.
const (
	codeOAuthAuthorizationPending = "08-03-3-001-200001"
	codeOAuthSlowDown             = "08-03-3-001-200002"
	codeOAuthAccessDenied         = "08-03-3-001-200003"
	codeOAuthExpiredToken         = "08-03-3-001-200004"
	codeOAuthInvalidGrant         = "08-03-3-001-200005"
	codeOAuthInvalidClient        = "08-03-3-001-200006"
	codeOAuthUnauthorizedClient   = "08-03-3-001-200007"
	codeOAuthInvalidScope         = "08-03-3-001-200008"
	codeOAuthUnsupportedGrant     = "08-03-3-001-200009"
	codeOAuthInvalidRequest       = "08-03-3-001-200010"
	codeOAuthServerError          = "08-03-3-001-200011"
)

// classifyDevicePollStatus maps poll failures by resultCode first (platform stable codes).
// resultMsg is only a legacy fallback for older environments.
func classifyDevicePollStatus(resultCode, resultMsg string) string {
	switch strings.TrimSpace(resultCode) {
	case codeOAuthAuthorizationPending:
		return pollStatusPending
	case codeOAuthSlowDown:
		return pollStatusSlowDown
	case codeOAuthAccessDenied:
		return pollStatusAccessDenied
	case codeOAuthExpiredToken:
		return pollStatusExpiredToken
	case codeOAuthInvalidGrant:
		return pollStatusInvalidGrant
	case codeOAuthInvalidClient:
		return pollStatusInvalidClient
	case codeOAuthUnauthorizedClient:
		return pollStatusUnauthorizedClient
	case codeOAuthInvalidScope:
		return pollStatusInvalidScope
	case codeOAuthUnsupportedGrant:
		return pollStatusUnsupportedGrant
	case codeOAuthInvalidRequest:
		return pollStatusInvalidRequest
	case codeOAuthServerError:
		return pollStatusServerError
	}

	// Legacy: English resultMsg (older docs / gateways).
	switch strings.ToLower(strings.TrimSpace(resultMsg)) {
	case "authorization_pending", "oauth2_authorization_pending":
		return pollStatusPending
	case "slow_down", "oauth2_slow_down":
		return pollStatusSlowDown
	case "access_denied", "oauth2_access_denied":
		return pollStatusAccessDenied
	case "expired_token", "oauth2_expired_token":
		return pollStatusExpiredToken
	case "invalid_grant", "oauth2_invalid_grant":
		return pollStatusInvalidGrant
	}

	return pollStatusUnknown
}

// RefreshAccessToken exchanges refresh_token for a new token pair.
func (c *Client) RefreshAccessToken(refreshToken string) (*TokenResponse, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, &RequestError{
			APIError: APIError{
				Code:      "unauthorized",
				Message:   "缺少 refresh_token，请重新运行 zhizai auth login",
				Reason:    "missing_refresh_token",
				Retryable: false,
			},
		}
	}

	form := url.Values{}
	form.Set("grant_type", refreshGrantType)
	form.Set("client_id", config.OAuthClientID)
	form.Set("refresh_token", refreshToken)

	env, err := c.postOAuthForm("/oauth2/token", form)
	if err != nil {
		return nil, err
	}
	if env.ResultCode != "0" {
		return nil, &RequestError{
			APIError: APIError{
				Code:      env.ResultCode,
				Message:   env.ResultMsg,
				Reason:    "refresh_failed",
				Retryable: false,
			},
		}
	}
	return decodeTokenObject(env.ResultObject)
}

func decodeTokenObject(raw json.RawMessage) (*TokenResponse, error) {
	var token TokenResponse
	if err := json.Unmarshal(raw, &token); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("token response missing access_token")
	}
	if token.TokenType == "" {
		token.TokenType = "Bearer"
	}
	return &token, nil
}

func (c *Client) postOAuthForm(path string, form url.Values) (*apiEnvelope, error) {
	c.waitRateLimit()

	reqURL := c.resolveOAuthURL(path)
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, networkRequestError(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, networkRequestError(err)
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("parsing OAuth response (HTTP %d): %w; body=%s", resp.StatusCode, err, truncate(string(raw), 200))
	}
	env.Raw = string(raw)
	return &env, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// EnsureAccessToken refreshes the access token when expired or missing (if refresh_token exists).
func (c *Client) EnsureAccessToken() error {
	if c.authMode != config.AuthModeOAuth {
		return nil
	}
	if c.accessToken != "" && !c.tokenNearExpiry() {
		return nil
	}
	if c.refreshToken == "" {
		if c.accessToken == "" {
			return &RequestError{
				APIError: APIError{
					Code:      "unauthorized",
					Message:   "未登录，请运行 zhizai auth login",
					Reason:    "missing_credentials",
					Retryable: false,
				},
			}
		}
		return nil
	}

	token, err := c.RefreshAccessToken(c.refreshToken)
	if err != nil {
		return err
	}
	c.accessToken = token.AccessToken
	if token.RefreshToken != "" {
		c.refreshToken = token.RefreshToken
	}
	c.tokenExpiresAt = 0
	if token.ExpiresIn > 0 {
		c.tokenExpiresAt = time.Now().Unix() + int64(token.ExpiresIn)
	}
	return config.Get().UpdateOAuthTokens(token.AccessToken, token.RefreshToken, token.TokenType, token.Scope, int(token.ExpiresIn))
}

func (c *Client) tokenNearExpiry() bool {
	if c.tokenExpiresAt <= 0 {
		return false
	}
	return time.Now().Unix() >= c.tokenExpiresAt-60
}
