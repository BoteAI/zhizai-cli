package ask

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewAskCmd returns the ask command for template-based summarization.
func NewAskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ask [question]",
		Short: "基于笔记的动态模版问答与总结",
		Long:  "将实现 note.md 中的动态模版管线：查模版 → 查笔记 → 成文。",
		Example: `  zhizai ask "本周会议总结一下" -o json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("ask")
		},
	}
}
