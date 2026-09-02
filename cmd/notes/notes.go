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

	cmd := &cobra.Command{
		Use:   "notes",
		Args:  cobra.NoArgs,
		Short: "查看笔记列表",
		Example: `  zhizai notes
  zhizai notes --limit 10
  zhizai notes --all
  zhizai notes --title 会议 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			if all {
				return streamAll(cmd, c, title, noteType)
			}
			if page <= 0 {
				page = 1
			}
			if limit <= 0 {
				limit = 20
			}

			data, err := c.NoteList(client.NoteListParams{
				Title:    title,
				NoteType: noteType,
				PageNum:  page,
				PageSize: limit,
			})
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
	return cmd
}

func streamAll(cmd *cobra.Command, c *client.Client, title, noteType string) error {
	isJSON := output.Format() == "json"
	page := 1
	var allNotes []client.Note
	totalShown := 0

	if !isJSON {
		printHeader(cmd)
	}

	for {
		data, err := c.NoteList(client.NoteListParams{
			Title:    title,
			NoteType: noteType,
			PageNum:  page,
			PageSize: 20,
		})
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
