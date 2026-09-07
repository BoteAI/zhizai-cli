package config

import (
	"os"
	"testing"
)

func TestResolveAPIBaseURL_Presets(t *testing.T) {
	t.Setenv("ZHIZAI_API_URL", "")
	t.Setenv("ZHIZAI_OAUTH_URL", "")
	t.Setenv("ZHIZAI_ENV", "")

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
	_ = os.Unsetenv
}

func TestResolveOAuthBaseURL(t *testing.T) {
	t.Setenv("ZHIZAI_API_URL", "")
	t.Setenv("ZHIZAI_OAUTH_URL", "")
	t.Setenv("ZHIZAI_ENV", "")

	got := ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[DefaultEnv].OAuthBase {
		t.Fatalf("default oauth = %q, want %q", got, EnvPresets[DefaultEnv].OAuthBase)
	}

	t.Setenv("ZHIZAI_ENV", "prod")
	got = ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[EnvProd].OAuthBase {
		t.Fatalf("prod oauth = %q", got)
	}

	t.Setenv("ZHIZAI_ENV", "test")
	got = ResolveOAuthBaseURL(&Config{})
	if got != EnvPresets[EnvTest].OAuthBase {
		t.Fatalf("test oauth = %q, want %q", got, EnvPresets[EnvTest].OAuthBase)
	}

	// Custom API only → OAuth falls back to same base.
	t.Setenv("ZHIZAI_ENV", "")
	t.Setenv("ZHIZAI_API_URL", "https://custom.example/api/v1")
	got = ResolveOAuthBaseURL(&Config{})
	if got != "https://custom.example/api/v1" {
		t.Fatalf("custom api-only oauth fallback = %q", got)
	}

	t.Setenv("ZHIZAI_OAUTH_URL", "https://oauth.example/server/")
	got = ResolveOAuthBaseURL(&Config{})
	if got != "https://oauth.example/server" {
		t.Fatalf("ZHIZAI_OAUTH_URL should win, got %q", got)
	}
}

func TestNormalizeEnv(t *testing.T) {
	if NormalizeEnv("lingxi") != EnvDev {
		t.Fatal("lingxi should map to dev")
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

	t.Setenv("ZHIZAI_API_URL", "")
	t.Setenv("ZHIZAI_OAUTH_URL", "")
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
			"https://lingxi.iwhalecloud.com/LCDP-RECORD/api/v1",
			"/note/createNote",
			"https://lingxi.iwhalecloud.com/LCDP-RECORD/api/v1/note/createNote",
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

	got = JoinOAuthURL("https://openapi.zzjilu.com/api/v1/", "oauth2/token")
	want = "https://openapi.zzjilu.com/api/v1/oauth2/token"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestActiveEnvName_CustomWhenOAuthOverride(t *testing.T) {
	t.Setenv("ZHIZAI_API_URL", "")
	t.Setenv("ZHIZAI_OAUTH_URL", "https://oauth.example/server")
	t.Setenv("ZHIZAI_ENV", "prod")
	if ActiveEnvName(&Config{}) != "custom" {
		t.Fatal("oauth override should mark env custom")
	}
}
