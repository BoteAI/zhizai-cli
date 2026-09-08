package config

import (
	"os"
	"strings"
)

// OAuth client_id registered for zhizai-cli (device authorization).
const OAuthClientID = "zhizai_cli"

// Environment names for ZHIZAI_ENV / config.env.
const (
	EnvDev  = "dev"  // 兼容别名，解析为 test
	EnvTest = "test" // 测试（lingxi）
	EnvGray = "gray" // 灰度（:9001）
	EnvProd = "prod" // 生产
)

// Auth modes stored in config.json.
const (
	AuthModeAPIKey = "api_key"
	AuthModeOAuth  = "oauth"
)

// EnvEndpoints holds per-environment service bases.
//
// OpenAPI（业务）与 OAuth2（设备授权）通常分属不同 base：
//   - APIBase：.../api/v1  → /note/...
//   - OAuthBase：.../server → /oauth2/...（网关会省略 app/note 前缀）
type EnvEndpoints struct {
	APIBase   string // business OpenAPI root
	OAuthBase string // OAuth2 authorize/token root
	SiteURL   string // portal home (文档 / 诊断展示用)
}

// EnvPresets maps env name → API and OAuth bases.
//
// 环境对照（与平台运维约定一致）：
//
//	| 环境 | 官网 | OAuth base | OpenAPI base |
//	| test | https://lingxi.iwhalecloud.com/zzjl/ | .../zzjl/server | .../zzjl/api/v1 |
//	| gray | https://www.zzjilu.com:9001/pc/home | ...:9001/server | ...:9001/api/v1 |
//	| prod | https://www.zzjilu.com/pc/home | .../server | openapi.zzjilu.com/api/v1 |
//
// 快速切换（仅本地开发，需 export ZHIZAI_DEV=1）：
//  1. 环境变量 ZHIZAI_API_URL / ZHIZAI_OAUTH_URL（完整覆盖）
//  2. config.json 的 api_url / oauth_url
//  3. 环境变量 ZHIZAI_ENV 或 config.json 的 env（查本表）
//  4. DefaultEnv（prod）
//
// 未设置 ZHIZAI_DEV=1 时（发布包默认），始终使用 EnvProd，忽略上述切换项。
var EnvPresets = map[string]EnvEndpoints{
	EnvProd: {
		APIBase:   "https://openapi.zzjilu.com/api/v1",
		OAuthBase: "https://www.zzjilu.com/server",
		SiteURL:   "https://www.zzjilu.com/pc/home",
	},
	EnvGray: {
		// openapi.zzjilu.com:9001/api/v1 暂未开通，临时走官网同域 api/v1。
		APIBase:   "https://www.zzjilu.com:9001/api/v1",
		OAuthBase: "https://www.zzjilu.com:9001/server",
		SiteURL:   "https://www.zzjilu.com:9001/pc/home",
	},
	EnvTest: {
		APIBase:   "https://lingxi.iwhalecloud.com/zzjl/api/v1",
		OAuthBase: "https://lingxi.iwhalecloud.com/zzjl/server",
		SiteURL:   "https://lingxi.iwhalecloud.com/zzjl/",
	},
	// dev：历史别名，与 test 相同（lingxi）。
	EnvDev: {
		APIBase:   "https://lingxi.iwhalecloud.com/zzjl/api/v1",
		OAuthBase: "https://lingxi.iwhalecloud.com/zzjl/server",
		SiteURL:   "https://lingxi.iwhalecloud.com/zzjl/",
	},
}

// ServiceBaseURLs is the legacy alias of EnvPresets[].APIBase.
// Prefer EnvPresets / ResolveAPIBaseURL for new code.
var ServiceBaseURLs = map[string]string{
	EnvProd: EnvPresets[EnvProd].APIBase,
	EnvGray: EnvPresets[EnvGray].APIBase,
	EnvTest: EnvPresets[EnvTest].APIBase,
	EnvDev:  EnvPresets[EnvDev].APIBase,
}

// DefaultEnv is used when neither ZHIZAI_ENV nor config.env is set.
// Published builds always resolve to prod regardless of this constant when AllowEnvSwitch is false.
const DefaultEnv = EnvProd

// DefaultAPIBaseURL is the resolved default for the active DefaultEnv.
// Kept for backward-compatible imports; prefer ResolveAPIBaseURL.
var DefaultAPIBaseURL = EnvPresets[DefaultEnv].APIBase

// AllowEnvSwitch reports whether non-prod endpoint switching is enabled.
//
// Published / end-user binaries lock to production. Local developers must set:
//
//	export ZHIZAI_DEV=1
//
// then ZHIZAI_ENV / api_url / oauth_url take effect.
func AllowEnvSwitch() bool {
	return strings.TrimSpace(os.Getenv("ZHIZAI_DEV")) == "1"
}

// NormalizeEnv returns a known env name or DefaultEnv.
func NormalizeEnv(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case EnvProd, "production":
		return EnvProd
	case EnvGray, "staging", "canary", "pre", "preprod", "灰度":
		return EnvGray
	case EnvTest, "sit", "qa", "uat", "lingxi":
		return EnvTest
	case EnvDev, "development", "develop":
		return EnvTest
	default:
		return DefaultEnv
	}
}

func resolveEnvName(cfg *Config) string {
	if !AllowEnvSwitch() {
		return EnvProd
	}
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
	if !AllowEnvSwitch() {
		return trimBase(EnvPresets[EnvProd].APIBase)
	}
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
// When AllowEnvSwitch is false, always returns production OAuth base.
// Otherwise priority: ZHIZAI_OAUTH_URL → config.oauth_url → named-env preset OAuthBase。
//
// 仅覆盖业务 API（ZHIZAI_API_URL / api_url）时，OAuth 仍走当前命名环境的 OAuthBase，
// 因为各环境 OpenAPI 与 OAuth 分属不同主机。
func ResolveOAuthBaseURL(cfg *Config) string {
	if !AllowEnvSwitch() {
		return trimBase(EnvPresets[EnvProd].OAuthBase)
	}
	if v := trimBase(os.Getenv("ZHIZAI_OAUTH_URL")); v != "" {
		return v
	}
	if cfg != nil {
		if v := trimBase(cfg.OAuthURL); v != "" {
			return v
		}
	}

	envName := resolveEnvName(cfg)
	if ep, ok := EnvPresets[envName]; ok {
		return trimBase(ep.OAuthBase)
	}
	return trimBase(EnvPresets[DefaultEnv].OAuthBase)
}

// ResolveSiteURL returns the portal home URL for the active named env (empty if custom).
func ResolveSiteURL(cfg *Config) string {
	name := ActiveEnvName(cfg)
	if name == "custom" {
		return ""
	}
	if ep, ok := EnvPresets[name]; ok {
		return trimBase(ep.SiteURL)
	}
	return trimBase(EnvPresets[DefaultEnv].SiteURL)
}

// ActiveEnvName reports which named env is in effect when no absolute URL override is set.
func ActiveEnvName(cfg *Config) string {
	if !AllowEnvSwitch() {
		return EnvProd
	}
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
