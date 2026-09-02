package doctor

import (
	"fmt"
	"os"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/config"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/version"
	"github.com/spf13/cobra"
)

type check struct {
	Name     string `json:"name"`
	OK       bool   `json:"ok"`
	Required bool   `json:"required"`
	Message  string `json:"message,omitempty"`
}

type response struct {
	Success              bool    `json:"success"`
	DiagnosticsCompleted bool    `json:"diagnostics_completed"`
	Ready                bool    `json:"ready"`
	Status               string  `json:"status"`
	CLIVersion           string  `json:"cli_version"`
	Checks               []check `json:"checks"`
}

// NewDoctorCmd returns the doctor command.
func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "检查 CLI、认证与 API 连通性",
		Example: `  zhizai doctor
  zhizai doctor -o json`,
		Args: cobra.NoArgs,
		RunE: runDoctor,
	}
}

func runDoctor(c *cobra.Command, _ []string) error {
	ping := func() error { return client.New().Ping() }
	result := evaluate(config.Get().IsLoggedIn() || os.Getenv("ZHIZAI_REC_API_KEY") != "", ping)

	format := output.Format()
	if format == "json" {
		if err := output.WriteSuccessJSON(c.OutOrStdout(), result); err != nil {
			return err
		}
		if !result.Ready {
			return fmt.Errorf("存在阻断问题，请运行 zhizai auth login")
		}
		return nil
	}

	fmt.Fprintln(c.OutOrStdout(), "智在记录 CLI 诊断")
	fmt.Fprintf(c.OutOrStdout(), "版本: %s\n\n", version.Version)
	for _, item := range result.Checks {
		icon := "✓"
		if !item.OK {
			icon = "✗"
		}
		fmt.Fprintf(c.OutOrStdout(), "%s %s: %s\n", icon, item.Name, item.Message)
	}
	fmt.Fprintf(c.OutOrStdout(), "\n状态: %s\n", result.Status)
	if !result.Ready {
		return fmt.Errorf("存在阻断问题，请运行 zhizai auth login")
	}
	return nil
}

func evaluate(authOK bool, ping func() error) response {
	checks := []check{
		{Name: "cli", OK: true, Required: true, Message: "zhizai CLI 可执行"},
		{Name: "auth", OK: authOK, Required: true, Message: authMessage(authOK)},
	}

	apiOK := false
	apiMsg := "OpenAPI 探活失败"
	if !authOK {
		apiMsg = "跳过：未配置 API Key"
	} else if ping != nil && ping() == nil {
		apiOK = true
		apiMsg = "OpenAPI 连通正常"
	}
	checks = append(checks, check{
		Name: "api", OK: apiOK, Required: true, Message: apiMsg,
	})

	ready := requiredChecksReady(checks)
	status := "not_ready"
	if ready {
		status = "ready"
	}

	return response{
		Success:              ready,
		DiagnosticsCompleted: true,
		Ready:                ready,
		Status:               status,
		CLIVersion:           version.String(),
		Checks:               checks,
	}
}

func requiredChecksReady(checks []check) bool {
	for _, item := range checks {
		if item.Required && !item.OK {
			return false
		}
	}
	return true
}

func authMessage(ok bool) string {
	if ok {
		return "API Key 已配置"
	}
	return "未配置 API Key"
}
