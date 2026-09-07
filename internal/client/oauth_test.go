package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeviceAuthorizeAndPoll(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/device/authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Fatalf("content-type %q", ct)
		}
		_ = r.ParseForm()
		if r.Form.Get("client_id") != "zhizai_cli" {
			t.Fatalf("client_id=%q", r.Form.Get("client_id"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "成功",
			"resultObject": map[string]interface{}{
				"device_code":               "dev-code",
				"user_code":                 "ABCD-EFGH",
				"verification_uri":          "https://example.test/oauth/device",
				"verification_uri_complete": "https://example.test/oauth/device?user_code=ABCD-EFGH",
				"expires_in":                "300",
				"interval":                  "5",
			},
		})
	})

	polls := 0
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("grant_type") == "refresh_token" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"resultCode": "0",
				"resultMsg":  "成功",
				"resultObject": map[string]interface{}{
					"access_token":  "access-new",
					"token_type":    "Bearer",
					"expires_in":    3600,
					"refresh_token": "refresh-new",
					"scope":         "note.content.read",
				},
			})
			return
		}
		polls++
		if polls == 1 {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"resultCode":   "1",
				"resultMsg":    "authorization_pending",
				"resultObject": nil,
			})
			return
		}
		if polls == 2 {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"resultCode":   "08-03-3-001-200001",
				"resultMsg":    "用户尚未确认设备授权",
				"resultObject": nil,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "成功",
			"resultObject": map[string]interface{}{
				"access_token":  "access-1",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-1",
				"scope":         "note.content.read",
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewWithOptions(server.URL, "", server.Client())
	session, _, err := c.DeviceAuthorize("")
	if err != nil {
		t.Fatal(err)
	}
	if session.UserCode != "ABCD-EFGH" || session.DeviceCode != "dev-code" {
		t.Fatalf("%+v", session)
	}
	if session.Interval != 5 || session.ExpiresIn != 300 {
		t.Fatalf("flexInt parse failed: interval=%d expires=%d", session.Interval, session.ExpiresIn)
	}

	token, pending, _, err := c.PollDeviceToken(session.DeviceCode)
	if err != nil || token != nil || pending != "authorization_pending" {
		t.Fatalf("pending: token=%v pending=%q err=%v", token, pending, err)
	}
	token, pending, _, err = c.PollDeviceToken(session.DeviceCode)
	if err != nil || token != nil || pending != "authorization_pending" {
		t.Fatalf("chinese pending: token=%v pending=%q err=%v", token, pending, err)
	}
	token, pending, _, err = c.PollDeviceToken(session.DeviceCode)
	if err != nil || token == nil || token.AccessToken != "access-1" {
		t.Fatalf("success: token=%v pending=%q err=%v", token, pending, err)
	}

	refreshed, err := c.RefreshAccessToken("refresh-1")
	if err != nil || refreshed.AccessToken != "access-new" || refreshed.RefreshToken != "refresh-new" {
		t.Fatalf("refresh: %+v err=%v", refreshed, err)
	}
}

func TestBusinessRequestUsesOAuthHeader(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/note/queryNoteList", func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-OAuth2-Access-Token")
		if got != "Bearer tok-abc" {
			t.Fatalf("oauth header = %q", got)
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatal("should not set Authorization in oauth mode")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resultCode": "0",
			"resultMsg":  "ok",
			"resultObject": map[string]interface{}{
				"pageNum": 1, "pageSize": 1, "total": "0", "hasNextPage": false, "list": []any{},
			},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewOAuthWithOptions(server.URL, "tok-abc", "ref-1", server.Client())
	if _, err := c.NoteList(NoteListParams{PageNum: 1, PageSize: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestClassifyDevicePollStatusByResultCode(t *testing.T) {
	cases := map[string]string{
		"08-03-3-001-200001": pollStatusPending,
		"08-03-3-001-200002": pollStatusSlowDown,
		"08-03-3-001-200003": pollStatusAccessDenied,
		"08-03-3-001-200004": pollStatusExpiredToken,
		"08-03-3-001-200005": pollStatusInvalidGrant,
		"08-03-3-001-200006": pollStatusInvalidClient,
		"08-03-3-001-200007": pollStatusUnauthorizedClient,
		"08-03-3-001-200008": pollStatusInvalidScope,
		"08-03-3-001-200009": pollStatusUnsupportedGrant,
		"08-03-3-001-200010": pollStatusInvalidRequest,
		"08-03-3-001-200011": pollStatusServerError,
	}
	for code, want := range cases {
		// resultMsg must not override resultCode
		got := classifyDevicePollStatus(code, "任意中文提示")
		if got != want {
			t.Fatalf("code %s: got %q want %q", code, got, want)
		}
	}
	if got := classifyDevicePollStatus("1", "authorization_pending"); got != pollStatusPending {
		t.Fatalf("legacy msg fallback: %q", got)
	}
	if got := classifyDevicePollStatus("1", "未知"); got != pollStatusUnknown {
		t.Fatalf("unknown: %q", got)
	}
}
