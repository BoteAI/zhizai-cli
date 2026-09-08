package config

import (
	"testing"
)

func enableDevSwitch(t *testing.T) {
	t.Helper()
	t.Setenv("ZHIZAI_DEV", "1")
	t.Setenv("ZHIZAI_API_URL", "")
	t.Setenv("ZHIZAI_OAUTH_URL", "")
	t.Setenv("ZHIZAI_ENV", "")
}

func TestPublishedLocksToProd(t *testing.T) {
	t.Setenv("ZHIZAI_DEV", "")
	t.Setenv("ZHIZAI_ENV", "test")
	t.Setenv("ZHIZAI_API_URL", "https://evil.example/api")
	t.Setenv("ZHIZAI_OAUTH_URL", "https://evil.example/oauth")

	if AllowEnvSwitch() {
		t.Fatal("AllowEnvSwitch should be false without ZHIZAI_DEV=1")
	}
	if got := ResolveAPIBaseURL(&Config{Env: EnvTest, APIURL: "https://cfg.example"}); got != EnvPresets[EnvProd].APIBase {
		t.Fatalf("published API = %q", got)
	}
	if got := ResolveOAuthBaseURL(&Config{Env: EnvDev, OAuthURL: "https://cfg.example"}); got != EnvPresets[EnvProd].OAuthBase {
		t.Fatalf("published OAuth = %q", got)
	}
	if ActiveEnvName(&Config{Env: EnvTest}) != EnvProd {
		t.Fatal("published ActiveEnvName must be prod")
	}
}

func TestResolveAPIBaseURL_Presets(t *testing.T) {
	enableDevSwitch(t)

	got := ResolveAPIBaseURL(&Config{})
	want := EnvPresets[DefaultEnv].APIBase
	if got != want {
		t.Fatalf("default = %q, want %q", got, want)
	}

	t.Setenv("ZHIZAI_ENV", "prod")
	got = ResolveAPIBaseURL(&Config{})
	if got != EnvPresets[EnvProd].APIBase {
		t.Fatalf("prod env = %q", got)
	}

	t.Setenv("ZHIZAI_ENV", "gray")
	got = ResolveAPIBaseURL(&Config{})
	if got != EnvPresets[EnvGray].APIBase {
		t.Fatalf("gray env = %q, want %q", got, EnvPresets[EnvGray].APIBase)
	}

	t.Setenv("ZHIZAI_ENV", "dev")
	got = ResolveAPIBaseURL(&Config{APIURL: "https://custom.example/api"})
	if got != "https://custom.example/api" {
		t.Fatalf("config api_url should win when no ZHIZAI_API_URL, got %q", got)
	}

	t.Setenv("ZHIZAI_API_URL", "https://override.example/v1/")
	got = ResolveAPIBaseURL(&Config{APIURL: "https://custom.example/api", Env: EnvProd})
	if got != "https://override.example/v1" {
		t.Fatalf("ZHIZAI_API_URL should win, got %q", got)
	}
}

func TestResolveOAuthBaseURL(t *testing.T) {
	enableDevSwitch(t)

	got := ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[DefaultEnv].OAuthBase {
		t.Fatalf("default oauth = %q, want %q", got, EnvPresets[DefaultEnv].OAuthBase)
	}

	t.Setenv("ZHIZAI_ENV", "prod")
	got = ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[EnvProd].OAuthBase {
		t.Fatalf("prod oauth = %q, want %q", got, EnvPresets[EnvProd].OAuthBase)
	}
	if got != "https://www.zzjilu.com/server" {
		t.Fatalf("prod oauth must be www.zzjilu.com/server, got %q", got)
	}

	t.Setenv("ZHIZAI_ENV", "test")
	got = ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[EnvTest].OAuthBase {
		t.Fatalf("test oauth = %q, want %q", got, EnvPresets[EnvTest].OAuthBase)
	}

	t.Setenv("ZHIZAI_ENV", "gray")
	got = ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[EnvGray].OAuthBase {
		t.Fatalf("gray oauth = %q, want %q", got, EnvPresets[EnvGray].OAuthBase)
	}

	// Custom API only → OAuth still follows named env (split hosts).
	t.Setenv("ZHIZAI_ENV", "test")
	t.Setenv("ZHIZAI_API_URL", "https://custom.example/api/v1")
	got = ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[EnvTest].OAuthBase {
		t.Fatalf("api-only override should keep named-env oauth, got %q", got)
	}

	t.Setenv("ZHIZAI_OAUTH_URL", "https://oauth.example/server/")
	got = ResolveOAuthBaseURL(&Config{})
	if got != "https://oauth.example/server" {
		t.Fatalf("ZHIZAI_OAUTH_URL should win, got %q", got)
	}
}

func TestNormalizeEnv(t *testing.T) {
	if NormalizeEnv("lingxi") != EnvTest {
		t.Fatal("lingxi should map to test")
	}
	if NormalizeEnv("production") != EnvProd {
		t.Fatal("production should map to prod")
	}
	if NormalizeEnv("test") != EnvTest {
		t.Fatal("test should map to test")
	}
	if NormalizeEnv("sit") != EnvTest {
		t.Fatal("sit should map to test")
	}
	if NormalizeEnv("gray") != EnvGray {
		t.Fatal("gray should map to gray")
	}
	if NormalizeEnv("staging") != EnvGray {
		t.Fatal("staging should map to gray")
	}
	if NormalizeEnv("dev") != EnvTest {
		t.Fatal("dev should alias to test")
	}

	enableDevSwitch(t)
	t.Setenv("ZHIZAI_ENV", "test")
	got := ResolveAPIBaseURL(&Config{})
	if got != EnvPresets[EnvTest].APIBase {
		t.Fatalf("test env = %q, want %q", got, EnvPresets[EnvTest].APIBase)
	}
}

func TestJoinAPIURL(t *testing.T) {
	cases := []struct {
		base, path, want string
	}{
		{
			"https://openapi.zzjilu.com/api/v1",
			"/note/queryNoteList",
			"https://openapi.zzjilu.com/api/v1/note/queryNoteList",
		},
		{
			"https://www.zzjilu.com:9001/api/v1",
			"/note/queryNoteList",
			"https://www.zzjilu.com:9001/api/v1/note/queryNoteList",
		},
		{
			"https://lingxi.iwhalecloud.com/zzjl/api/v1",
			"/note/queryNoteList",
			"https://lingxi.iwhalecloud.com/zzjl/api/v1/note/queryNoteList",
		},
		{
			"http://10.10.179.140:8058/portal/lcdp-app/server/app/note",
			"/note/queryNoteList",
			"http://10.10.179.140:8058/portal/lcdp-app/server/app/note/queryNoteList",
		},
		{
			"http://10.10.179.140:8058/portal/lcdp-app/server/app/note/",
			"note/querySingleNoteDetail?noteId=1",
			"http://10.10.179.140:8058/portal/lcdp-app/server/app/note/querySingleNoteDetail?noteId=1",
		},
	}
	for _, tc := range cases {
		got := JoinAPIURL(tc.base, tc.path)
		if got != tc.want {
			t.Fatalf("JoinAPIURL(%q, %q) = %q, want %q", tc.base, tc.path, got, tc.want)
		}
	}
}

func TestJoinOAuthURL(t *testing.T) {
	got := JoinOAuthURL("https://lingxi.iwhalecloud.com/zzjl/server", "/oauth2/device/confirm")
	want := "https://lingxi.iwhalecloud.com/zzjl/server/oauth2/device/confirm"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	got = JoinOAuthURL("https://www.zzjilu.com/server/", "oauth2/token")
	want = "https://www.zzjilu.com/server/oauth2/token"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	got = JoinOAuthURL("https://www.zzjilu.com:9001/server", "/oauth2/device/authorize")
	want = "https://www.zzjilu.com:9001/server/oauth2/device/authorize"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestActiveEnvName_CustomWhenOAuthOverride(t *testing.T) {
	enableDevSwitch(t)
	t.Setenv("ZHIZAI_OAUTH_URL", "https://oauth.example/server")
	t.Setenv("ZHIZAI_ENV", "prod")
	if ActiveEnvName(&Config{}) != "custom" {
		t.Fatal("oauth override should mark env custom")
	}
}

func TestResolveSiteURL(t *testing.T) {
	enableDevSwitch(t)
	t.Setenv("ZHIZAI_ENV", "gray")
	if got := ResolveSiteURL(&Config{}); got != "https://www.zzjilu.com:9001/pc/home" {
		t.Fatalf("gray site = %q", got)
	}
}
