package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/config"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/platform"
	"github.com/spf13/cobra"
)

var agentNames = map[string]string{
	"codex":          "codex",
	"claude-code":    "claude-code",
	"cursor":         "cursor",
	"gemini-cli":     "gemini-cli",
	"github-copilot": "github-copilot",
	"windsurf":       "windsurf",
	"opencode":       "opencode",
	"cline":          "cline",
}

var platformNames = map[string]string{
	"workbuddy": "WorkBuddy", "codex": "Codex", "claude-code": "Claude Code", "cursor": "Cursor",
	"openclaw": "OpenClaw（小龙虾）", "qclaw": "QClaw",
	"gemini-cli": "Gemini CLI", "github-copilot": "GitHub Copilot",
	"windsurf": "Windsurf", "opencode": "OpenCode", "cline": "Cline",
}

var marketplaceTargets = map[string]bool{"qclaw": true, "openclaw": true}

var platformPriority = map[string]int{
	"openclaw": 0, "claude-code": 1, "codex": 2, "cursor": 3,
	"github-copilot": 4, "gemini-cli": 5, "opencode": 6, "windsurf": 7,
	"cline": 8, "qclaw": 9, "workbuddy": 10,
}

const clawHubURL = "https://clawhub.ai"

const setupBanner = `
 ███████╗██╗  ██╗██╗██╗     ██╗     ███████╗
 ██╔════╝██║ ██╔╝██║██║     ██║     ██╔════╝
 ███████╗█████╔╝ ██║██║     ██║     ███████╗
 ╚════██║██╔═██╗ ██║██║     ██║     ╚════██║
 ███████║██║  ██╗██║███████╗███████╗███████║
 ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚══════╝`

const defaultCLIPackage = "@zhizai/cli@latest"

var workBuddySkillNames = []string{
	"zhizai-auth",
	"zhizai-note",
	"zhizai-knowledge",
	"zhizai-scene",
	"zhizai-team",
	"zhizai-msg",
}

type result struct {
	Success         bool             `json:"success"`
	Targets         []string         `json:"targets"`
	InstalledCLI    bool             `json:"installed_cli"`
	InstalledSkills bool             `json:"installed_skills"`
	RestartRequired []string         `json:"restart_required,omitempty"`
	Authenticated   bool             `json:"authenticated"`
	Platforms       []platformResult `json:"platforms"`
	NextActions     []nextAction     `json:"next_actions"`
	Next            string           `json:"next,omitempty"`
}

type platformResult struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	SkillsInstalled bool   `json:"skills_installed"`
	RestartRequired bool   `json:"restart_required"`
	Message         string `json:"message"`
}

type nextAction struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
}

func runInstaller(command *exec.Cmd) (string, error) {
	var out bytes.Buffer
	command.Stdout = &out
	command.Stderr = &out
	err := command.Run()
	return strings.TrimSpace(out.String()), err
}

func installerError(label string, err error, details string) error {
	if details == "" {
		return fmt.Errorf("%s失败: %w", label, err)
	}
	return fmt.Errorf("%s失败: %w\n%s", label, err, details)
}

func writeProgress(cmd *cobra.Command, outFormat, message string) {
	if outFormat == "table" {
		fmt.Fprintln(cmd.OutOrStdout(), message)
	}
}

func configureInstallProcess(command *exec.Cmd, outFormat string, stdout, stderr interface{ Write([]byte) (int, error) }) {
	command.Stdout = stdout
	command.Stderr = stderr
	if outFormat == "json" {
		command.Stdout = stderr
	}
}

// NewSetupCmd installs bundled atomic skills into supported local AI hosts.
func NewSetupCmd() *cobra.Command {
	var targets []string
	var scope, source string
	var skipAuth, dryRun, skipCLIInstall bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "安装 Skill 到本机 AI 并引导授权",
		Example: `  zhizai setup
  zhizai setup --dry-run -o json
  zhizai setup --skill-source . --skip-cli-install --skip-auth`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if scope != "global" && scope != "project" {
				return fmt.Errorf("不支持的安装范围: %s", scope)
			}
			resolved, err := resolveTargets(targets)
			if err != nil {
				return err
			}
			outFormat := output.Format()
			if dryRun {
				platforms, actions := setupPlatformResults(resolved, false)
				return writeResult(cmd, outFormat, result{
					Success:       true,
					Targets:       resolved,
					Authenticated: config.Get().IsLoggedIn() || os.Getenv("ZHIZAI_REC_API_KEY") != "",
					Platforms:     platforms,
					NextActions:   actions,
					Next:          setupPlan(resolved, scope),
				})
			}

			writeProgress(cmd, outFormat, setupBanner)
			if skipCLIInstall {
				writeProgress(cmd, outFormat, "\n正在同步智在记录 Skills，请稍候…")
			} else {
				writeProgress(cmd, outFormat, "\n正在安装智在记录，请稍候…")
			}

			if !skipCLIInstall {
				installCLI := exec.Command("npm", "install", "-g", cliPackage())
				installCLI.Stdin = cmd.InOrStdin()
				details, installErr := runInstaller(installCLI)
				if installErr != nil {
					return installerError("安装命令行工具", installErr, details)
				}
			}

			localTargets := locallyManagedTargets(resolved)
			if source == "" && len(localTargets) > 0 {
				source, err = globalPackageDir()
				if err != nil {
					return err
				}
			}

			agentTargets := standardAgentTargets(resolved)
			if len(agentTargets) > 0 {
				writeProgress(cmd, outFormat, "正在为 "+displayNames(resolved, false)+" 安装 Skills…")
				installArgs := []string{"-y", "skills", "add", source, "-y"}
				if scope == "global" {
					installArgs = append(installArgs, "-g")
				}
				installArgs = append(installArgs, "--agent")
				installArgs = append(installArgs, agentTargets...)
				install := exec.Command("npx", installArgs...)
				install.Stdin = cmd.InOrStdin()
				details, installErr := runInstaller(install)
				if installErr != nil {
					return installerError("安装 AI Skills", installErr, details)
				}
			}

			restartRequired := []string{}
			if contains(resolved, "workbuddy") {
				writeProgress(cmd, outFormat, "正在为 WorkBuddy 安装 Skills…")
				if err := installWorkBuddySkills(filepath.Join(source, "skills"), workBuddySkillsDir()); err != nil {
					return fmt.Errorf("安装 WorkBuddy Skills 失败: %w", err)
				}
				restartRequired = append(restartRequired, "workbuddy")
			}

			authed := config.Get().IsLoggedIn() || os.Getenv("ZHIZAI_REC_API_KEY") != ""
			if !skipAuth && !authed {
				writeProgress(cmd, outFormat, "接下来请配置智在记录 API Key…")
				login := exec.Command(os.Args[0], "auth", "login")
				login.Stdin = cmd.InOrStdin()
				configureInstallProcess(login, outFormat, cmd.OutOrStdout(), cmd.ErrOrStderr())
				if err := login.Run(); err != nil {
					return fmt.Errorf("智在记录授权失败: %w", err)
				}
				authed = true
			}

			platforms, actions := setupPlatformResults(resolved, true)
			next := "说“帮我列出最近笔记”完成验证"
			if !authed {
				next = "运行 zhizai auth login --api-key <key> 完成授权"
			}
			return writeResult(cmd, outFormat, result{
				Success:         true,
				Targets:         resolved,
				InstalledCLI:    !skipCLIInstall,
				InstalledSkills: len(localTargets) > 0,
				RestartRequired: restartRequired,
				Authenticated:   authed,
				Platforms:       platforms,
				NextActions:     actions,
				Next:            next,
			})
		},
	}

	cmd.Flags().StringSliceVar(&targets, "target", nil, "通常无需填写；仅用于调试时指定平台 ID")
	cmd.Flags().StringVar(&scope, "scope", "global", "Skill 安装范围: global 或 project")
	cmd.Flags().StringVar(&source, "skill-source", "", "Skill 来源；默认使用全局 @zhizai/cli 内置 Skills，本地验收可传仓库目录")
	cmd.Flags().BoolVar(&skipAuth, "skip-auth", false, "跳过首次授权")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "仅输出将执行的操作")
	cmd.Flags().BoolVar(&skipCLIInstall, "skip-cli-install", false, "跳过 CLI 安装，仅同步 Skills")
	_ = cmd.Flags().MarkHidden("skip-cli-install")
	return cmd
}

func setupPlan(targets []string, scope string) string {
	steps := []string{"npm install -g " + cliPackage()}
	agents := standardAgentTargets(targets)
	if len(agents) > 0 {
		args := []string{"npx -y skills add <全局 @zhizai/cli 目录> -y"}
		if scope == "global" {
			args = append(args, "-g")
		}
		steps = append(steps, strings.Join(args, " ")+" --agent "+strings.Join(agents, " "))
	}
	if contains(targets, "workbuddy") {
		steps = append(steps, "复制 Skills 到 "+workBuddySkillsDir()+" 并重启 WorkBuddy")
	}
	for _, target := range targets {
		if marketplaceTargets[target] {
			steps = append(steps, "在 "+platformNames[target]+" 内确认智在记录 Skill 已启用")
		}
	}
	return strings.Join(steps, " && ")
}

func locallyManagedTargets(targets []string) []string {
	result := []string{}
	for _, target := range targets {
		if target == "workbuddy" || agentNames[target] != "" {
			result = append(result, target)
		}
	}
	return result
}

func setupPlatformResults(targets []string, installed bool) ([]platformResult, []nextAction) {
	platforms := []platformResult{}
	actions := []nextAction{}
	skillCount := fmt.Sprintf("%d", len(workBuddySkillNames))
	for _, target := range targets {
		name := platformNames[target]
		switch {
		case marketplaceTargets[target]:
			platforms = append(platforms, platformResult{
				ID: target, Name: name, Status: "verify_in_platform",
				Message: "由平台管理 Skill，请在技能市场确认“智在记录”已安装并启用",
			})
			actions = append(actions, nextAction{
				ID: "verify_" + target + "_skill",
				Description: "在 " + name + " 内确认“智在记录”Skill 已安装并启用",
				URL: clawHubURL,
			})
		case target == "workbuddy":
			status, message := "planned", "将安装 "+skillCount+" 个 Skills；完成后需要重启 WorkBuddy"
			if installed {
				status, message = "installed", skillCount+" 个 Skills 已安装，重启 WorkBuddy 后生效"
			}
			platforms = append(platforms, platformResult{
				ID: target, Name: name, Status: status,
				SkillsInstalled: installed, RestartRequired: true, Message: message,
			})
		default:
			status, message := "planned", "将安装 "+skillCount+" 个 Skills"
			if installed {
				status, message = "installed", skillCount+" 个 Skills 已安装"
			}
			platforms = append(platforms, platformResult{
				ID: target, Name: name, Status: status,
				SkillsInstalled: installed, Message: message,
			})
		}
	}
	if len(targets) == 0 {
		actions = append(actions, nextAction{
			ID: "open_supported_ai",
			Description: "打开支持的 AI 应用后重新运行 zhizai setup",
		})
	}
	return platforms, actions
}

func cliPackage() string {
	if configured := strings.TrimSpace(os.Getenv("ZHIZAI_CLI_PACKAGE")); configured != "" {
		return configured
	}
	return defaultCLIPackage
}

func standardAgentTargets(targets []string) []string {
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		if agent, ok := agentNames[target]; ok {
			result = append(result, agent)
		}
	}
	return result
}

func displayNames(targets []string, includeMarketplace bool) string {
	names := []string{}
	for _, target := range targets {
		if target == "workbuddy" || (!includeMarketplace && marketplaceTargets[target]) {
			continue
		}
		names = append(names, platformNames[target])
	}
	return strings.Join(names, "、")
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func globalPackageDir() (string, error) {
	out, err := exec.Command("npm", "root", "-g").Output()
	if err != nil {
		return "", fmt.Errorf("无法定位全局 npm 目录: %w", err)
	}
	dir := filepath.Join(strings.TrimSpace(string(out)), "@zhizai", "cli")
	if info, err := os.Stat(filepath.Join(dir, "skills")); err != nil || !info.IsDir() {
		return "", fmt.Errorf("全局智在记录 CLI 缺少内置 Skills: %s", dir)
	}
	return dir, nil
}

func workBuddySkillsDir() string {
	if configured := strings.TrimSpace(os.Getenv("ZHIZAI_WORKBUDDY_SKILLS_DIR")); configured != "" {
		return configured
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".workbuddy", "skills")
	}
	return filepath.Join(home, ".workbuddy", "skills")
}

func installWorkBuddySkills(sourceRoot, targetRoot string) error {
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return err
	}
	stagingRoot, err := os.MkdirTemp(targetRoot, ".zhizai-skills-install-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stagingRoot)

	for _, name := range workBuddySkillNames {
		source := filepath.Join(sourceRoot, name)
		if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
			return fmt.Errorf("%s 缺少 SKILL.md", source)
		}
		if err := copyDir(source, filepath.Join(stagingRoot, name)); err != nil {
			return err
		}
	}

	backupRoot, err := os.MkdirTemp(targetRoot, ".zhizai-skills-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(backupRoot)
	committed := []string{}
	rollback := func() {
		for index := len(committed) - 1; index >= 0; index-- {
			name := committed[index]
			target := filepath.Join(targetRoot, name)
			_ = os.RemoveAll(target)
			backup := filepath.Join(backupRoot, name)
			if _, err := os.Stat(backup); err == nil {
				_ = os.Rename(backup, target)
			}
		}
	}
	for _, name := range workBuddySkillNames {
		target := filepath.Join(targetRoot, name)
		backup := filepath.Join(backupRoot, name)
		if _, err := os.Stat(target); err == nil {
			if err := os.Rename(target, backup); err != nil {
				rollback()
				return err
			}
		} else if !os.IsNotExist(err) {
			rollback()
			return err
		}
		if err := os.Rename(filepath.Join(stagingRoot, name), target); err != nil {
			if _, backupErr := os.Stat(backup); backupErr == nil {
				_ = os.Rename(backup, target)
			}
			rollback()
			return err
		}
		committed = append(committed, name)
	}
	return nil
}

func copyDir(source, target string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if info.IsDir() {
			return os.MkdirAll(destination, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("不支持的 Skill 文件类型: %s", path)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		outputFile, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			_ = input.Close()
			return err
		}
		_, copyErr := io.Copy(outputFile, input)
		inputCloseErr := input.Close()
		closeErr := outputFile.Close()
		if copyErr != nil {
			return copyErr
		}
		if inputCloseErr != nil {
			return inputCloseErr
		}
		return closeErr
	})
}

func resolveTargets(values []string) ([]string, error) {
	set := map[string]bool{}
	for _, value := range values {
		for _, id := range strings.Split(value, ",") {
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			if _, ok := platformNames[id]; !ok {
				return nil, fmt.Errorf("不支持的平台: %s", id)
			}
			set[id] = true
		}
	}
	if len(set) == 0 {
		for _, item := range platform.Detect() {
			if item.Detected {
				if _, ok := platformNames[item.ID]; ok {
					set[item.ID] = true
				}
			}
		}
	}
	result := make([]string, 0, len(set))
	for id := range set {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool {
		left, leftOK := platformPriority[result[i]]
		right, rightOK := platformPriority[result[j]]
		if leftOK && rightOK && left != right {
			return left < right
		}
		if leftOK != rightOK {
			return leftOK
		}
		return result[i] < result[j]
	})
	return result, nil
}

func writeResult(cmd *cobra.Command, outFormat string, data result) error {
	if outFormat == "json" {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "\n安装完成")
	if data.InstalledCLI {
		fmt.Fprintln(cmd.OutOrStdout(), "✓ 智在记录命令行工具")
	}
	if data.Authenticated {
		fmt.Fprintln(cmd.OutOrStdout(), "✓ 智在记录账号已连接")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "→ 智在记录账号尚未连接")
	}
	if len(data.Platforms) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "✓ 命令行工具已准备好")
		fmt.Fprintln(cmd.OutOrStdout(), "未检测到正在使用的 AI 应用；打开 AI 应用后再次运行这条安装命令即可")
	}
	for _, item := range data.Platforms {
		if item.Status == "installed" {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ %s：%s\n", item.Name, item.Message)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "→ %s：%s\n", item.Name, item.Message)
		}
	}
	if len(data.NextActions) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\n还需完成：")
		for _, action := range data.NextActions {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s", action.Description)
			if action.URL != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "：%s", action.URL)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
	}
	if data.Next != "" {
		fmt.Fprintln(cmd.OutOrStdout(), data.Next)
	}
	return nil
}
