package doctor

import "testing"

func TestAuthMessage(t *testing.T) {
	if got := authMessage(true); got != "API Key 已配置" {
		t.Fatalf("authMessage(true) = %q", got)
	}
	if got := authMessage(false); got != "未配置 API Key" {
		t.Fatalf("authMessage(false) = %q", got)
	}
}

func TestRequiredChecksReady(t *testing.T) {
	ok := []check{
		{Name: "cli", OK: true, Required: true},
		{Name: "auth", OK: true, Required: true},
		{Name: "api", OK: true, Required: true},
	}
	if !requiredChecksReady(ok) {
		t.Fatal("all required checks ok should be ready")
	}

	blocked := []check{
		{Name: "cli", OK: true, Required: true},
		{Name: "auth", OK: false, Required: true},
	}
	if requiredChecksReady(blocked) {
		t.Fatal("failed required check must block readiness")
	}

	optionalFail := []check{
		{Name: "cli", OK: true, Required: true},
		{Name: "warn", OK: false, Required: false},
	}
	if !requiredChecksReady(optionalFail) {
		t.Fatal("optional failure must not block readiness")
	}
}

func TestEvaluateWithoutAuth(t *testing.T) {
	result := evaluate(false, func() error { return nil })
	if result.Ready || result.Status != "not_ready" {
		t.Fatalf("expected not ready, got %+v", result)
	}
	if len(result.Checks) != 3 {
		t.Fatalf("checks = %d", len(result.Checks))
	}
	if result.Checks[1].OK || result.Checks[2].OK {
		t.Fatalf("auth/api should fail without credentials: %+v", result.Checks)
	}
}

func TestEvaluateWithAuthAndPing(t *testing.T) {
	result := evaluate(true, func() error { return nil })
	if !result.Ready || result.Status != "ready" {
		t.Fatalf("expected ready, got %+v", result)
	}
	if !result.Success || !result.DiagnosticsCompleted {
		t.Fatalf("success flags: %+v", result)
	}
}

func TestEvaluateWithAuthButPingFails(t *testing.T) {
	result := evaluate(true, func() error { return fmtError("boom") })
	if result.Ready {
		t.Fatalf("ping failure should block readiness: %+v", result)
	}
	if result.Checks[2].OK || result.Checks[2].Message != "OpenAPI 探活失败" {
		t.Fatalf("api check = %+v", result.Checks[2])
	}
}

type fmtError string

func (e fmtError) Error() string { return string(e) }
