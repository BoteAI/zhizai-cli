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
					"auth login":       "网页设备授权登录并探活（支持 --no-open）",
					"auth status":      "查看认证状态",
					"auth logout":      "清除本机凭证",
					"doctor":           "检查 CLI、认证与 API",
					"notes":            "笔记列表（支持摘要/总结/正文/时间/短链过滤）",
					"note get":         "笔记详情（支持 --short-url / --with-append）",
					"note create":      "创建文字笔记",
					"note update":      "更新笔记标题/摘要/总结",
					"note delete":      "删除笔记（需 --yes）",
					"note status":      "笔记处理进度",
					"note wait":        "等待笔记处理至终态",
					"note audio":       "下载笔记录音（--out 指定路径）",
					"file upload":      "上传文件，返回 fileId（不创建笔记）",
					"file get":         "按 fileId 下载文件（--out 指定路径）",
					"save":             "自动判型做成笔记（内部上传+创建，异步默认等待）",
					"knowledge list":   "笔记集列表",
					"knowledge get":    "笔记集详情",
					"knowledge notes":  "笔记集内笔记",
					"setup":            "安装 Skill 到本机 AI",
					"ask":              "问小智：默认自动建会话后问答",
					"ask session":      "仅创建问小智会话",
					"ask chat":         "仅流式问答（需 chatId/contextId）",
					"ask refs":         "查询问小智参考文档",
					"ask template":     "动态成文模版（可选附带笔记）",
					"summarize":        "文字流式总结（SSE）",
					"recordings":       "按标题/时间查总结与录音",
					"cards":            "我创建的知识卡分页列表",
				},
				Guarantees: map[string]any{
					"ids_as_strings":     true,
					"rate_limit_2rps":    true,
					"no_bearer_prefix":   true,
					"oauth_login":        true,
					"auth_login_no_open": true,
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
