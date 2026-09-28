package config

import (
	"testing"
)

// 未设置 ZHIZAI_DEV=1 时，无论是否 dev 构建都必须锁定生产。
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
