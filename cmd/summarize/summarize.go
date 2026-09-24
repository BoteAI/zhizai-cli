package summarize

import (
	"fmt"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewSummarizeCmd returns the text summary SSE command.
func NewSummarizeCmd() *cobra.Command {
	var content string
	var sceneID string

	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "对一段文字做流式总结（非问小智）",
		Example: `  zhizai summarize --content "明天开会讨论渠道" -o json
  zhizai summarize --content "..." --scene-id <sceneId> -o json`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			content = strings.TrimSpace(content)
			if content == "" {
				return fmt.Errorf("content 不能为空（请传 --content）")
			}
			c := client.New()
			text, err := c.CreateTextNoteSummary(client.TextNoteSummaryParams{
				Content: content,
				SceneID: strings.TrimSpace(sceneID),
			})
			if err != nil {
				return err
			}
			data := map[string]any{
				"summary": text,
				"sceneId": strings.TrimSpace(sceneID),
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			fmt.Fprintln(cmd.OutOrStdout(), text)
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "待总结正文")
	cmd.Flags().StringVar(&sceneID, "scene-id", "", "场景 ID（点名场景时先查真实 id）")
	return cmd
}
