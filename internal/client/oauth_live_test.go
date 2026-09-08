package client

import (
	"os"
	"testing"

	"github.com/BoteAI/zhizai-cli/internal/config"
)

func TestLiveLingxiDeviceAuthorize(t *testing.T) {
	if os.Getenv("ZHIZAI_LIVE") != "1" {
		t.Skip("set ZHIZAI_LIVE=1 to hit lingxi")
	}
	os.Setenv("ZHIZAI_DEV", "1")
	os.Setenv("ZHIZAI_ENV", "test")
	os.Unsetenv("ZHIZAI_API_URL")
	os.Unsetenv("ZHIZAI_OAUTH_URL")
	config.ResetForTests()
	c := New()
	s, _, err := c.DeviceAuthorize("")
	if err != nil {
		t.Fatal(err)
	}
	if s.UserCode == "" || s.VerificationURIComplete == "" {
		t.Fatalf("%+v", s)
	}
	t.Logf("user_code=%s interval=%d expires=%d url=%s", s.UserCode, s.Interval, s.ExpiresIn, s.VerificationURIComplete)
}
