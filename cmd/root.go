package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/BoteAI/zhizai-cli/cmd/ask"
	"github.com/BoteAI/zhizai-cli/cmd/auth"
	"github.com/BoteAI/zhizai-cli/cmd/capabilities"
	"github.com/BoteAI/zhizai-cli/cmd/doctor"
	"github.com/BoteAI/zhizai-cli/cmd/file"
	"github.com/BoteAI/zhizai-cli/cmd/knowledge"
	"github.com/BoteAI/zhizai-cli/cmd/msg"
	"github.com/BoteAI/zhizai-cli/cmd/note"
	"github.com/BoteAI/zhizai-cli/cmd/notes"
	"github.com/BoteAI/zhizai-cli/cmd/scene"
	"github.com/BoteAI/zhizai-cli/cmd/setup"
	"github.com/BoteAI/zhizai-cli/cmd/team"
	"github.com/BoteAI/zhizai-cli/cmd/update"
	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/config"
	clioutput "github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	apiKey       string
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:     "zhizai",
	Short:   "智在记录命令行工具 / CLI for 智在记录",
	Version: version.Version,
	Long: `zhizai 是智在记录的命令行工具，支持查询、管理笔记与知识库。
适合人工操作和 AI Agent 集成使用。

zhizai is a command-line tool for 智在记录 OpenAPI.
It allows both humans and AI agents to manage notes from the terminal.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		writeError(os.Stderr, err, outputFormat)
		os.Exit(1)
	}
}

func writeError(w io.Writer, err error, format string) {
	var requestErr *client.RequestError
	if format == "json" {
		apiErr := client.APIError{
			Code:      "cli_error",
			Message:   err.Error(),
			Reason:    "cli_error",
			Retryable: false,
		}
		if errors.As(err, &requestErr) {
			apiErr = requestErr.APIError
		}
		payload := clioutput.JSONResponse{
			Success: false,
			Data:    nil,
			Error:   apiErr,
		}
		if encodeErr := clioutput.WriteJSON(w, payload); encodeErr == nil {
			return
		}
	}
	fmt.Fprintln(w, err)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key（覆盖配置与 ZHIZAI_REC_API_KEY 环境变量）")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "输出格式: table 或 json")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if outputFormat != "table" && outputFormat != "json" {
			return fmt.Errorf("不支持的输出格式 %q；可用值: table, json", outputFormat)
		}
		clioutput.SetFormat(outputFormat)
		return nil
	}

	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(capabilities.NewCapabilitiesCmd())
	rootCmd.AddCommand(doctor.NewDoctorCmd())
	rootCmd.AddCommand(setup.NewSetupCmd())
	rootCmd.AddCommand(update.NewUpdateCmd())
	rootCmd.AddCommand(notes.NewNotesCmd())
	rootCmd.AddCommand(note.NewNoteCmd())
	rootCmd.AddCommand(file.NewFileCmd())
	rootCmd.AddCommand(ask.NewAskCmd())
	rootCmd.AddCommand(scene.NewSceneCmd())
	rootCmd.AddCommand(knowledge.NewKnowledgeCmd())
	rootCmd.AddCommand(team.NewTeamCmd())
	rootCmd.AddCommand(msg.NewMsgCmd())
	rootCmd.AddCommand(newVersionCmd())
}

func initConfig() {
	cfg := config.Get()

	if apiKey != "" {
		cfg.APIKey = apiKey
		return
	}

	if envKey := os.Getenv("ZHIZAI_REC_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}
}
