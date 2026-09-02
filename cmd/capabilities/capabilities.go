package capabilities

import (
	"fmt"

	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/version"
	"github.com/spf13/cobra"
)

type response struct {
	Success         bool              `json:"success"`
	CLIVersion      string            `json:"cli_version"`
	ContractVersion string            `json:"contract_version"`
	Commands        map[string]string `json:"commands"`
	Guarantees      map[string]any    `json:"guarantees"`
}

// NewCapabilitiesCmd returns the capabilities command.
func NewCapabilitiesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "查看 CLI 稳定能力契约",
		Example: `  zhizai capabilities -o json`,
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			data := response{
				Success:         true,
				CLIVersion:      version.String(),
				ContractVersion: "0.1.0",
				Commands: map[string]string{
					"auth login":  "保存 API Key 并探活",
					"auth status": "查看认证状态",
					"auth logout": "清除本机凭证",
					"doctor":      "检查 CLI、认证与 API",
					"notes":       "笔记列表",
					"note get":    "笔记详情",
					"setup":       "安装 Skill 到本机 AI",
					"file":        "文件上传（待实现）",
					"ask":         "动态模版总结（待实现）",
				},
				Guarantees: map[string]any{
					"ids_as_strings":     true,
					"rate_limit_2rps":    true,
					"no_bearer_prefix":   true,
					"oauth_login":        false,
					"json_output_flag":   "-o json",
					"config_path":        "~/.zhizai/config.json",
					"env_api_key":        "ZHIZAI_REC_API_KEY",
					"default_base_url":   "https://openapi.zzjilu.com/api/v1",
				},
			}

			if output.Format() == "json" {
				return output.WriteSuccessJSON(c.OutOrStdout(), data)
			}

			fmt.Fprintf(c.OutOrStdout(), "zhizai capabilities (contract %s)\n", data.ContractVersion)
			for name, desc := range data.Commands {
				fmt.Fprintf(c.OutOrStdout(), "  %-12s %s\n", name, desc)
			}
			return nil
		},
	}
}
