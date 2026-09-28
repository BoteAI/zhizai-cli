//go:build !dev

package config

import (
	"testing"
)

// 发布构建（无 dev tag）的核心保障：即使用户机器上残留
// ZHIZAI_DEV=1 / ZHIZAI_ENV=test / api_url 等全部切换项，
// 也必须锁定生产环境——切换能力在编译期已被移除。
func TestReleaseBuildIgnoresDevFlag(t *testing.T) {
	t.Setenv("ZHIZAI_DEV", "1")
	t.Setenv("ZHIZAI_ENV", "test")
	t.Setenv("ZHIZAI_API_URL", "https://evil.example/api")
	t.Setenv("ZHIZAI_OAUTH_URL", "https://evil.example/oauth")

	if AllowEnvSwitch() {
		t.Fatal("release build must ignore ZHIZAI_DEV=1 (built without -tags dev)")
	}
	cfg := &Config{Env: EnvTest, APIURL: "https://cfg.example/api", OAuthURL: "https://cfg.example/oauth"}
	if got := ResolveAPIBaseURL(cfg); got != EnvPresets[EnvProd].APIBase {
		t.Fatalf("release API = %q, want prod %q", got, EnvPresets[EnvProd].APIBase)
	}
	if got := ResolveOAuthBaseURL(cfg); got != EnvPresets[EnvProd].OAuthBase {
		t.Fatalf("release OAuth = %q, want prod %q", got, EnvPresets[EnvProd].OAuthBase)
	}
	if got := ActiveEnvName(cfg); got != EnvProd {
		t.Fatalf("release ActiveEnvName = %q, want prod", got)
	}
	if got := ResolveSiteURL(cfg); got != EnvPresets[EnvProd].SiteURL {
		t.Fatalf("release SiteURL = %q, want prod %q", got, EnvPresets[EnvProd].SiteURL)
	}
}
