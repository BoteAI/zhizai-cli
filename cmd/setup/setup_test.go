package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTargetsRejectsUnsupported(t *testing.T) {
	if _, err := resolveTargets([]string{"not-a-platform"}); err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

func TestResolveTargetsUsesPriorityOrder(t *testing.T) {
	got, err := resolveTargets([]string{"workbuddy,cursor,openclaw"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"openclaw", "cursor", "workbuddy"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestStandardAgentTargets(t *testing.T) {
	got := standardAgentTargets([]string{"cursor", "openclaw", "codex"})
	if len(got) != 2 || got[0] != "cursor" || got[1] != "codex" {
		t.Fatalf("got %v", got)
	}
}

func TestSetupPlatformResults(t *testing.T) {
	platforms, actions := setupPlatformResults([]string{"cursor", "qclaw", "workbuddy"}, true)
	if platforms[0].Status != "installed" || !platforms[0].SkillsInstalled {
		t.Fatalf("cursor = %+v", platforms[0])
	}
	if platforms[1].Status != "verify_in_platform" || len(actions) == 0 {
		t.Fatalf("qclaw = %+v actions=%+v", platforms[1], actions)
	}
	if platforms[2].Status != "installed" || !platforms[2].RestartRequired {
		t.Fatalf("workbuddy = %+v", platforms[2])
	}
}

func TestDryRunDoesNotClaimInstalled(t *testing.T) {
	platforms, _ := setupPlatformResults([]string{"cursor"}, false)
	if platforms[0].Status == "installed" || platforms[0].SkillsInstalled {
		t.Fatalf("planned state claimed installed: %+v", platforms[0])
	}
}

func TestInstallWorkBuddySkills(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	for _, name := range workBuddySkillNames {
		dir := filepath.Join(source, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := installWorkBuddySkills(source, target); err != nil {
		t.Fatal(err)
	}
	for _, name := range workBuddySkillNames {
		if _, err := os.Stat(filepath.Join(target, name, "SKILL.md")); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestCLIPackageOverride(t *testing.T) {
	t.Setenv("ZHIZAI_CLI_PACKAGE", "@zhizai/cli@0.1.0-test")
	if got := cliPackage(); got != "@zhizai/cli@0.1.0-test" {
		t.Fatalf("cliPackage() = %q", got)
	}
}

func TestParseVersionOutput(t *testing.T) {
	got, ok := parseVersionOutput("zhizai version 0.0.6\n")
	if !ok || got != "0.0.6" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
	if _, ok := parseVersionOutput("something else"); ok {
		t.Fatal("expected false")
	}
}

func TestProbeCLIReadyMissingBinary(t *testing.T) {
	ready, _ := probeCLIReady(filepath.Join(t.TempDir(), "zhizai"), func(string) (string, error) {
		t.Fatal("should not run version")
		return "", nil
	})
	if ready {
		t.Fatal("missing binary must not be ready")
	}
}

func TestShouldInstallCLIForceWinsOverSkip(t *testing.T) {
	if !shouldInstallCLI(true, true, true) {
		t.Fatal("force must win over skip and ready")
	}
}

func TestShouldInstallCLI(t *testing.T) {
	cases := []struct {
		force, skip, ready bool
		want               bool
	}{
		{false, true, false, false},
		{false, false, true, false},
		{false, false, false, true},
		{true, false, true, true},
	}
	for _, tc := range cases {
		if got := shouldInstallCLI(tc.force, tc.skip, tc.ready); got != tc.want {
			t.Fatalf("force=%v skip=%v ready=%v: got %v want %v", tc.force, tc.skip, tc.ready, got, tc.want)
		}
	}
}

func TestSetupPlanForceWinsOverSkip(t *testing.T) {
	plan := setupPlan(nil, "global", true, true, true)
	if !strings.Contains(plan, "npm install -g") {
		t.Fatalf("expected install step, got %q", plan)
	}
	if strings.Contains(plan, "跳过 npm install") {
		t.Fatalf("must not skip install when force is set: %q", plan)
	}
}

func TestProbeCLIReadyOK(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "zhizai")
	if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	ready, ver := probeCLIReady(bin, func(string) (string, error) {
		return "zhizai version 0.0.6\n", nil
	})
	if !ready || ver != "0.0.6" {
		t.Fatalf("ready=%v ver=%q", ready, ver)
	}
}
