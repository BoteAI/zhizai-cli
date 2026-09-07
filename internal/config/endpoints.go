package config

import (
	"os"
	"strings"
)

// OAuth client_id registered for zhizai-cli (device authorization).
const OAuthClientID = "zhizai_cli"

// Environment names for ZHIZAI_ENV / config.env.
const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

// Auth modes stored in config.json.
const (
	AuthModeAPIKey = "api_key"
	AuthModeOAuth  = "oauth"
)

// EnvEndpoints holds per-environment service bases.
//
// Production keeps a single OpenAPI root for both business and OAuth APIs.
// Test/dev split hosts: business may be LCDP (.../app/note) or OpenAPI (.../api/v1),
// while device OAuth lives under the zzjl server root.
type EnvEndpoints struct {
	APIBase   string // business OpenAPI / LCDP root
	OAuthBase string // OAuth2 token/authorize root (.../oauth2/...)
}

// EnvPresets maps env name → API and OAuth bases.
//
// 快速切换方式（优先级从高到低）：
//  1. 环境变量 ZHIZAI_API_URL / ZHIZAI_OAUTH_URL（完整覆盖）
//  2. config.json 的 api_url / oauth_url
//  3. 环境变量 ZHIZAI_ENV 或 config.json 的 env（查本表）
//  4. DefaultEnv
//
// 业务相对路径固定为 /note/...；若 APIBase 已以 /note 结尾（LCDP），
// JoinAPIURL 会去掉重复的 /note 前缀。
var EnvPresets = map[string]EnvEndpoints{
	EnvProd: {
		APIBase:   "https://openapi.zzjilu.com/api/v1",
		OAuthBase: "https://openapi.zzjilu.com/api/v1",
	},
	EnvDev: {
		APIBase:   "https://lingxi.iwhalecloud.com/LCDP-RECORD/api/v1",
		OAuthBase: "https://lingxi.iwhalecloud.com/zzjl/server",
	},
	EnvTest: {
		APIBase:   "https://lingxi.iwhalecloud.com/zzjl/api/v1",
		OAuthBase: "https://lingxi.iwhalecloud.com/zzjl/server",
	},
}

// ServiceBaseURLs is the legacy alias of EnvPresets[].APIBase.
// Prefer EnvPresets / ResolveAPIBaseURL for new code.
var ServiceBaseURLs = map[string]string{
	EnvProd: EnvPresets[EnvProd].APIBase,
	EnvDev:  EnvPresets[EnvDev].APIBase,
	EnvTest: EnvPresets[EnvTest].APIBase,
}

// DefaultEnv is used when neither ZHIZAI_ENV nor config.env is set.
const DefaultEnv = EnvDev

// DefaultAPIBaseURL is the resolved default for the active DefaultEnv.
// Kept for backward-compatible imports; prefer ResolveAPIBaseURL.
var DefaultAPIBaseURL = EnvPresets[DefaultEnv].APIBase

// NormalizeEnv returns a known env name or DefaultEnv.
func NormalizeEnv(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case EnvProd, "production":
		return EnvProd
	case EnvTest, "sit", "qa", "uat":
		return EnvTest
	case EnvDev, "development", "develop", "lingxi":
		return EnvDev
	default:
		return DefaultEnv
	}
}

func resolveEnvName(cfg *Config) string {
	if v := strings.TrimSpace(os.Getenv("ZHIZAI_ENV")); v != "" {
		return NormalizeEnv(v)
	}
	if cfg != nil && cfg.Env != "" {
		return NormalizeEnv(cfg.Env)
	}
	return DefaultEnv
}

func trimBase(v string) string {
	return strings.TrimRight(strings.TrimSpace(v), "/")
}

// ResolveAPIBaseURL returns the business API base URL without trailing slash.
func ResolveAPIBaseURL(cfg *Config) string {
	if v := trimBase(os.Getenv("ZHIZAI_API_URL")); v != "" {
		return v
	}
	if cfg != nil {
		if v := trimBase(cfg.APIURL); v != "" {
			return v
		}
	}
	envName := resolveEnvName(cfg)
	if ep, ok := EnvPresets[envName]; ok {
		return trimBase(ep.APIBase)
	}
	return trimBase(EnvPresets[DefaultEnv].APIBase)
}

// ResolveOAuthBaseURL returns the OAuth2 API base URL without trailing slash.
//
// Priority: ZHIZAI_OAUTH_URL → config.oauth_url → named-env preset →
// if only a custom api_url/ZHIZAI_API_URL is set, fall back to that API base
// (single-base mode, matches production).
func ResolveOAuthBaseURL(cfg *Config) string {
	if v := trimBase(os.Getenv("ZHIZAI_OAUTH_URL")); v != "" {
		return v
	}
	if cfg != nil {
		if v := trimBase(cfg.OAuthURL); v != "" {
			return v
		}
	}

	apiOverride := trimBase(os.Getenv("ZHIZAI_API_URL"))
	if apiOverride == "" && cfg != nil {
		apiOverride = trimBase(cfg.APIURL)
	}

	// Named env without absolute URL override → use preset OAuth base.
	if apiOverride == "" {
		envName := resolveEnvName(cfg)
		if ep, ok := EnvPresets[envName]; ok {
			return trimBase(ep.OAuthBase)
		}
		return trimBase(EnvPresets[DefaultEnv].OAuthBase)
	}

	// Custom API base only: keep OAuth on the same root (prod-compatible).
	return apiOverride
}

// ActiveEnvName reports which named env is in effect when no absolute URL override is set.
func ActiveEnvName(cfg *Config) string {
	if strings.TrimSpace(os.Getenv("ZHIZAI_API_URL")) != "" || strings.TrimSpace(os.Getenv("ZHIZAI_OAUTH_URL")) != "" {
		return "custom"
	}
	if cfg != nil && (strings.TrimSpace(cfg.APIURL) != "" || strings.TrimSpace(cfg.OAuthURL) != "") {
		return "custom"
	}
	return resolveEnvName(cfg)
}

// JoinAPIURL joins a business base with a relative API path such as /note/queryNoteList.
//
// If base already ends with "/note" (LCDP style) and path starts with "/note/",
// the duplicate "/note" segment is removed so:
//
//	base=.../server/app/note + /note/queryNoteList → .../server/app/note/queryNoteList
//
// OpenAPI style bases ending in /api/v1 keep the /note prefix:
//
//	base=.../api/v1 + /note/queryNoteList → .../api/v1/note/queryNoteList
func JoinAPIURL(base, path string) string {
	base = trimBase(base)
	path = "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	if strings.HasSuffix(base, "/note") && (path == "/note" || strings.HasPrefix(path, "/note/")) {
		path = strings.TrimPrefix(path, "/note")
		if path == "" {
			path = "/"
		}
	}
	return base + path
}

// JoinOAuthURL joins an OAuth base with a relative path such as /oauth2/token.
func JoinOAuthURL(base, path string) string {
	base = trimBase(base)
	path = "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	return base + path
}
