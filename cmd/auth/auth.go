package auth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewAuthCmd returns the auth command tree.
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "管理认证",
		Long:  "默认通过网页设备授权登录；也可刷新令牌或查看状态。凭证保存在本机 ~/.zhizai/config.json。",
		Example: `  zhizai auth login
  zhizai auth refresh
  zhizai auth status
  zhizai auth logout`,
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newRefreshCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newStatusCmd())
	return cmd
}

func newLoginCmd() *cobra.Command {
	var key string
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "login",
		Args:  cobra.NoArgs,
		Short: "登录（网页设备授权）",
		Long:  `通过 OAuth 2.0 设备授权登录：打开浏览器确认后，CLI 轮询获取 access_token。`,
		Example: `  zhizai auth login
  zhizai auth login --no-open`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			printServiceEndpoint(out)

			if cmd.Flags().Changed("api-key") {
				return loginWithAPIKey(out, key)
			}
			return runDeviceFlow(out, noOpen)
		},
	}

	cmd.Flags().BoolVar(&noOpen, "no-open", false, "不自动打开浏览器；打印授权 URL 供宿主展示")
	cmd.Flags().StringVar(&key, "api-key", "", "API Key（仅网页授权失败时的备用方式）")
	_ = cmd.Flags().MarkHidden("api-key")
	return cmd
}

func printServiceEndpoint(out io.Writer) {
	if !config.AllowEnvSwitch() {
		return
	}
	cfg := config.Get()
	apiBase := config.ResolveAPIBaseURL(cfg)
	oauthBase := config.ResolveOAuthBaseURL(cfg)
	env := config.ActiveEnvName(cfg)
	fmt.Fprintf(out, "服务环境: %s（开发模式 ZHIZAI_DEV=1）\n业务基址: %s\n", env, apiBase)
	if oauthBase == apiBase {
		fmt.Fprintf(out, "OAuth基址: %s（与业务共用）\n", oauthBase)
	} else {
		fmt.Fprintf(out, "OAuth基址: %s\n", oauthBase)
	}
}

func apiKeyFallbackHint() string {
	return "\n\n若网页授权失败，可改用 API Key：\n  zhizai auth login --api-key <your-api-key>"
}

func wrapAuthFailure(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w%s", err, apiKeyFallbackHint())
}

func loginWithAPIKey(out io.Writer, key string) error {
	if key == "" {
		fmt.Fprint(out, "请粘贴 API Key（来自 https://www.zzjilu.com/pc/developer ）: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading API key: %w", err)
		}
		key = strings.TrimSpace(line)
	}
	if err := validateAPIKey(key); err != nil {
		return err
	}
	key = strings.TrimSpace(key)

	if err := config.Get().SetAPIKeyLogin(key); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if err := client.New().Ping(); err != nil {
		return fmt.Errorf("API Key 已保存，但连接验证失败: %w", err)
	}

	fmt.Fprintln(out, "✅ 已使用 API Key 登录并验证连接成功。")
	return nil
}

func runDeviceFlow(out io.Writer, noOpen bool) error {
	c := client.New()
	session, rawAuth, err := c.DeviceAuthorize("")
	if err != nil {
		if config.AllowEnvSwitch() && rawAuth != "" {
			fmt.Fprintf(out, "[oauth] device/authorize 原始返回:\n%s\n", rawAuth)
		}
		return wrapAuthFailure(fmt.Errorf("创建设备授权会话失败: %w", err))
	}
	if config.AllowEnvSwitch() {
		fmt.Fprintf(out, "[oauth] device/authorize 原始返回:\n%s\n", rawAuth)
	}

	openURL := session.VerificationURIComplete
	if openURL == "" {
		openURL = session.VerificationURI
	}

	if noOpen {
		// Machine-readable lines for WorkBuddy / headless hosts.
		fmt.Fprintf(out, "[device_code] %s\n", session.UserCode)
		fmt.Fprintf(out, "[verify_url]  %s\n", openURL)
		fmt.Fprintf(out, "[expires_in]  %d\n", int(session.ExpiresIn))
		fmt.Fprintf(out, "[interval]    %d\n", int(session.Interval))
	} else {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "请在浏览器打开以下地址完成授权：\n\n  %s\n\n", openURL)
		fmt.Fprintf(out, "确认码: %s\n\n", session.UserCode)
		if openURL != "" {
			openBrowser(openURL)
		} else {
			fmt.Fprintf(out, "请手动打开 %s 并输入确认码 %s\n\n", session.VerificationURI, session.UserCode)
		}
		fmt.Fprintf(out, "等待授权确认（最长约 %d 秒，每 %d 秒轮询一次）\n", int(session.ExpiresIn), int(session.Interval))
	}

	interval := time.Duration(session.Interval) * time.Second
	deadline := time.Now().Add(time.Duration(session.ExpiresIn) * time.Second)

	// Spec: wait at least interval before first poll.
	time.Sleep(interval)

	pollN := 0
	for time.Now().Before(deadline) {
		pollN++
		token, pending, rawPoll, err := c.PollDeviceToken(session.DeviceCode)
		if config.AllowEnvSwitch() {
			fmt.Fprintf(out, "\n[oauth] token 轮询 #%d 原始返回:\n%s\n", pollN, maskTokenJSON(rawPoll))
		}
		if err != nil {
			var fail error
			switch pending {
			case "access_denied":
				fail = fmt.Errorf("用户拒绝了授权")
			case "expired_token", "invalid_grant":
				fail = fmt.Errorf("授权会话已失效（%s），请重新运行 zhizai auth login", pending)
			case "invalid_client":
				fail = fmt.Errorf("客户端配置无效（%s），请检查 client_id", pending)
			case "unauthorized_client":
				fail = fmt.Errorf("客户端不允许设备授权（%s），请检查授权模式", pending)
			case "invalid_scope":
				fail = fmt.Errorf("请求 scope 不在授权范围内（%s）", pending)
			case "unsupported_grant_type":
				fail = fmt.Errorf("grant_type 不受支持（%s），请升级 CLI", pending)
			case "invalid_request":
				fail = fmt.Errorf("请求参数不符合要求（%s）", pending)
			case "server_error":
				fail = fmt.Errorf("服务端错误（%s），请稍后重试或联系平台", pending)
			default:
				fail = fmt.Errorf("轮询 token 失败: %w", err)
			}
			return wrapAuthFailure(fail)
		}
		if token != nil {
			if err := config.Get().SetOAuthLogin(token.AccessToken, token.RefreshToken, token.TokenType, token.Scope, int(token.ExpiresIn)); err != nil {
				return wrapAuthFailure(fmt.Errorf("saving tokens: %w", err))
			}
			if err := client.New().Ping(); err != nil {
				if !noOpen {
					fmt.Fprintf(out, "⚠️ 令牌已保存，但探活失败: %v\n", err)
					fmt.Fprintln(out, "✅ 网页授权登录成功（探活失败可稍后重试业务命令）。")
				}
				return nil
			}
			if !noOpen {
				fmt.Fprintln(out, "✅ 网页授权登录成功。")
			}
			return nil
		}

		if config.AllowEnvSwitch() {
			fmt.Fprintf(out, "[oauth] 状态=%s，继续等待...\n", pending)
		} else if !noOpen {
			fmt.Fprint(out, ".")
		}
		wait := interval
		if pending == "slow_down" {
			wait = interval + 5*time.Second
		}
		time.Sleep(wait)
	}

	if !noOpen {
		fmt.Fprintln(out)
	}
	return wrapAuthFailure(fmt.Errorf("授权超时，请重新运行 zhizai auth login"))
}

// maskTokenJSON redacts access_token / refresh_token in raw JSON for console logs.
func maskTokenJSON(raw string) string {
	if raw == "" {
		return raw
	}
	out := raw
	for _, key := range []string{"access_token", "refresh_token"} {
		// naive redact: "key":"...value..."
		prefix := `"` + key + `":"`
		for {
			i := strings.Index(out, prefix)
			if i < 0 {
				break
			}
			start := i + len(prefix)
			end := strings.Index(out[start:], `"`)
			if end < 0 {
				break
			}
			end += start
			val := out[start:end]
			masked := maskKey(val)
			out = out[:start] + masked + out[end:]
			// advance past this occurrence
			out = out[:start+len(masked)+1] + out[start+len(masked)+1:]
			break
		}
	}
	return out
}

func newRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "refresh",
		Args:    cobra.NoArgs,
		Short:   "刷新 OAuth access_token",
		Example: `  zhizai auth refresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()
			if cfg.AuthMode != config.AuthModeOAuth || cfg.RefreshToken == "" {
				return fmt.Errorf("当前不是 OAuth 登录或缺少 refresh_token，请先 zhizai auth login")
			}
			printServiceEndpoint(cmd.OutOrStdout())
			c := client.New()
			token, err := c.RefreshAccessToken(cfg.RefreshToken)
			if err != nil {
				return err
			}
			if err := cfg.UpdateOAuthTokens(token.AccessToken, token.RefreshToken, token.TokenType, token.Scope, int(token.ExpiresIn)); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "✅ access_token 已刷新。")
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Args:    cobra.NoArgs,
		Short:   "清除本机保存的凭证",
		Example: `  zhizai auth logout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Get().Clear(); err != nil {
				return fmt.Errorf("clearing config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "已退出登录。")
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Args:    cobra.NoArgs,
		Short:   "查看当前认证状态",
		Example: `  zhizai auth status`,
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			printServiceEndpoint(out)

			if envKey := os.Getenv("ZHIZAI_REC_API_KEY"); envKey != "" {
				fmt.Fprintln(out, "认证方式: 环境变量 ZHIZAI_REC_API_KEY")
				fmt.Fprintf(out, "API Key: %s\n", maskKey(envKey))
				return
			}

			cfg := config.Get()
			if !cfg.IsLoggedIn() {
				fmt.Fprintln(out, "未认证。请运行: zhizai auth login")
				return
			}

			switch cfg.AuthMode {
			case config.AuthModeOAuth:
				fmt.Fprintln(out, "认证方式: OAuth 设备授权")
				fmt.Fprintf(out, "access_token: %s\n", maskKey(cfg.AccessToken))
				if cfg.RefreshToken != "" {
					fmt.Fprintf(out, "refresh_token: %s\n", maskKey(cfg.RefreshToken))
				}
				if cfg.Scope != "" {
					fmt.Fprintf(out, "scope: %s\n", cfg.Scope)
				}
				if cfg.TokenExpiresAt > 0 {
					exp := time.Unix(cfg.TokenExpiresAt, 0).Local()
					fmt.Fprintf(out, "过期时间: %s\n", exp.Format("2006年1月2日 15:04:05"))
					if cfg.TokenExpired(0) {
						fmt.Fprintln(out, "状态: access_token 已过期（将自动 refresh 或请运行 zhizai auth refresh）")
					}
				}
			default:
				fmt.Fprintln(out, "认证方式: API Key")
				fmt.Fprintf(out, "API Key: %s\n", maskKey(cfg.APIKey))
				if cfg.ExpiresAt != "" {
					fmt.Fprintf(out, "过期时间: %s\n", cfg.ExpiresAt)
				}
			}
		},
	}
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}

func validateAPIKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("API Key 不能为空")
	}
	return nil
}

func openBrowser(rawURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "linux":
		cmd = exec.Command("xdg-open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		return
	}
	_ = cmd.Start()
}
