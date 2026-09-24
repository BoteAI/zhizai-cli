package recordings

import (
	"fmt"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

const sep = "  "

var cols = []ui.ColSpec{
	{Value: "NoteID", Width: 18},
	{Value: "Title", Width: 36},
	{Value: "Status", Width: 12},
	{Value: "Created", Width: 19},
}

// NewRecordingsCmd returns the HR summary/recording query command.
func NewRecordingsCmd() *cobra.Command {
	var title, start, end, from, to string

	cmd := &cobra.Command{
		Use:   "recordings",
		Short: "按标题/时间查询总结与录音（人资）",
		Example: `  zhizai recordings --title 经营 --start "2026-09-01 00:00:00" --end "2026-09-30 23:59:59" -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			startTime := start
			endTime := end
			if from != "" {
				startTime = from
			}
			if to != "" {
				endTime = to
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/querySummaryAndRecording"))
			list, err := c.QuerySummaryAndRecording(client.SummaryAndRecordingParams{
				Title:     title,
				StartTime: startTime,
				EndTime:   endTime,
			})
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), list)
			}
			if len(list) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "没有符合条件的记录")
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(cols, sep))
			fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(cols, sep))
			for _, item := range list {
				row := []ui.ColSpec{
					{Value: item.NoteID, Width: cols[0].Width},
					{Value: item.NoteTitle, Width: cols[1].Width},
					{Value: item.NoteStatus, Width: cols[2].Width},
					{Value: item.NoteCreateTime, Width: cols[3].Width},
				}
				fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n(%d records)\n", len(list))
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "标题模糊匹配")
	cmd.Flags().StringVar(&start, "start", "", "开始时间")
	cmd.Flags().StringVar(&end, "end", "", "结束时间")
	cmd.Flags().StringVar(&from, "from", "", "开始时间（--start 别名）")
	cmd.Flags().StringVar(&to, "to", "", "结束时间（--end 别名）")
	return cmd
}
