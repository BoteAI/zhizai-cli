package notes

import (
	"fmt"
	"strconv"
	"time"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

const sep = "  "

var cols = []ui.ColSpec{
	{Value: "ID", Width: 18},
	{Value: "Title", Width: 36},
	{Value: "Type", Width: 10},
	{Value: "State", Width: 12},
	{Value: "Created", Width: 19},
}

// NewNotesCmd returns the notes list command.
func NewNotesCmd() *cobra.Command {
	var limit int
	var page int
	var all bool
	var title string
	var noteType string
	var abstract string
	var summary string
	var content string
	var start string
	var end string
	var from string
	var to string
	var withContent bool
	var withShortURL bool

	cmd := &cobra.Command{
		Use:   "notes",
		Args:  cobra.NoArgs,
		Short: "查看笔记列表",
		Example: `  zhizai notes
  zhizai notes --limit 10
  zhizai notes --all
  zhizai notes --title 会议 -o json
  zhizai notes --abstract 渠道 --from "2026-09-01 00:00:00" --to "2026-09-30 23:59:59"
  zhizai notes --with-content --with-short-url`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteList"))

			startTime := start
			endTime := end
			if from != "" {
				startTime = from
			}
			if to != "" {
				endTime = to
			}

			params := client.NoteListParams{
				Title:           title,
				AbstractContent: abstract,
				Summary:         summary,
				Content:         content,
				NoteType:        noteType,
				StartTime:       startTime,
				EndTime:         endTime,
			}
			if withContent {
				params.WithContent = "true"
			}
			if withShortURL {
				params.WithShortUrl = "true"
			}

			if all {
				return streamAll(cmd, c, params)
			}
			if page <= 0 {
				page = 1
			}
			if limit <= 0 {
				limit = 20
			}
			params.PageNum = page
			params.PageSize = limit

			data, err := c.NoteList(params)
			if err != nil {
				return err
			}

			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}

			printHeader(cmd)
			for _, n := range data.List {
				printRow(cmd, n)
			}
			total := data.Total
			if total == "" {
				total = strconv.Itoa(len(data.List))
			}
			if data.HasNextPage {
				fmt.Fprintf(cmd.OutOrStdout(),
					"\n(showing page %d, %d notes — use --page %d or --all)\n",
					data.PageNum, len(data.List), data.PageNum+1)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\n(%d notes, total %s)\n", len(data.List), total)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "每页数量")
	cmd.Flags().IntVar(&page, "page", 1, "页码")
	cmd.Flags().BoolVar(&all, "all", false, "自动翻页获取全部")
	cmd.Flags().StringVar(&title, "title", "", "按标题模糊筛选")
	cmd.Flags().StringVar(&noteType, "type", "", "按类型筛选: text/voice/document/link/image/knowCard")
	cmd.Flags().StringVar(&abstract, "abstract", "", "按摘要模糊筛选")
	cmd.Flags().StringVar(&summary, "summary", "", "按总结模糊筛选")
	cmd.Flags().StringVar(&content, "content", "", "按正文模糊筛选")
	cmd.Flags().StringVar(&start, "start", "", "开始时间（创建时间范围）")
	cmd.Flags().StringVar(&end, "end", "", "结束时间（创建时间范围）")
	cmd.Flags().StringVar(&from, "from", "", "开始时间（--start 别名）")
	cmd.Flags().StringVar(&to, "to", "", "结束时间（--end 别名）")
	cmd.Flags().BoolVar(&withContent, "with-content", false, "列表结果带正文")
	cmd.Flags().BoolVar(&withShortURL, "with-short-url", false, "列表结果带短链")
	return cmd
}

func streamAll(cmd *cobra.Command, c *client.Client, base client.NoteListParams) error {
	isJSON := output.Format() == "json"
	page := 1
	var allNotes []client.Note
	totalShown := 0

	if !isJSON {
		printHeader(cmd)
	}

	for {
		params := base
		params.PageNum = page
		params.PageSize = 20
		data, err := c.NoteList(params)
		if err != nil {
			return err
		}
		if isJSON {
			allNotes = append(allNotes, data.List...)
		} else {
			for _, n := range data.List {
				printRow(cmd, n)
				totalShown++
			}
		}
		if !data.HasNextPage || len(data.List) == 0 {
			break
		}
		page++
		time.Sleep(500 * time.Millisecond)
	}

	if isJSON {
		return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]interface{}{
			"list":        allNotes,
			"total":       strconv.Itoa(len(allNotes)),
			"hasNextPage": false,
			"pageNum":     1,
			"pageSize":    len(allNotes),
			"size":        len(allNotes),
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n(%d notes total)\n", totalShown)
	return nil
}

func printHeader(cmd *cobra.Command) {
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(cols, sep))
	fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(cols, sep))
}

func printRow(cmd *cobra.Command, n client.Note) {
	row := []ui.ColSpec{
		{Value: n.ID, Width: cols[0].Width},
		{Value: n.Title, Width: cols[1].Width},
		{Value: client.NoteTypeLabel(n.NoteType), Width: cols[2].Width},
		{Value: n.NoteState, Width: cols[3].Width},
		{Value: n.CreateTime, Width: cols[4].Width},
	}
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
}
