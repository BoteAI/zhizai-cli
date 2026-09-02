package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewAuthCmd returns the auth command tree.
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "管理认证",
		Long:  "登录、退出或检查智在记录 API Key。凭证只保存在本机 ~/.zhizai/config.json。",
		Example: `  zhizai auth login --api-key <key>
  zhizai auth status
  zhizai auth logout`,
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newStatusCmd())
	return cmd
}

func newLoginCmd() *cobra.Command {
	var key string

	cmd := &cobra.Command{
		Use:   "login",
		Args:  cobra.NoArgs,
		Short: "保存 API Key 并验证连接",
		Long:  "v1 支持直接保存 API Key。OAuth 浏览器授权将在后续版本提供。",
		Example: `  zhizai auth login --api-key <key>
  zhizai auth login`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				fmt.Fprint(cmd.OutOrStdout(), "请粘贴 API Key（来自 https://www.zzjilu.com/pc/developer ）: ")
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

			cfg := config.Get()
			cfg.APIKey = key
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if err := client.New().Ping(); err != nil {
				return fmt.Errorf("API Key 已保存，但连接验证失败: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✅ 已登录并验证连接成功。")
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "api-key", "", "API Key（跳过交互输入）")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Args:    cobra.NoArgs,
		Short:   "清除本机保存的 API Key",
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
			if envKey := os.Getenv("ZHIZAI_REC_API_KEY"); envKey != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "已通过环境变量 ZHIZAI_REC_API_KEY 认证。")
				return
			}

			cfg := config.Get()
			if cfg.IsLoggedIn() {
				fmt.Fprintf(cmd.OutOrStdout(), "已认证。API Key: %s\n", maskKey(cfg.APIKey))
				if cfg.ExpiresAt != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "过期时间: %s\n", cfg.ExpiresAt)
				}
				return
			}
			fmt.Fprintln(cmd.OutOrStdout(), "未认证。请运行: zhizai auth login")
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
