package cards

import (
	"fmt"
	"strconv"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

const sep = "  "

var cols = []ui.ColSpec{
	{Value: "ID", Width: 18},
	{Value: "Name", Width: 36},
	{Value: "Cards", Width: 8},
	{Value: "Created", Width: 19},
}

// NewCardsCmd returns the knowledge-cards list command.
func NewCardsCmd() *cobra.Command {
	var page, limit int

	cmd := &cobra.Command{
		Use:   "cards",
		Short: "分页查询我的知识卡笔记",
		Example: `  zhizai cards
  zhizai scene cards
  zhizai cards --page 1 --limit 10 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if page <= 0 {
				page = 1
			}
			if limit <= 0 {
				limit = 10
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryKnowledgeCardByPage"))
			data, err := c.KnowledgeCards(client.KnowledgeCardsParams{
				PageNum:  page,
				PageSize: limit,
			})
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			if len(data.List) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "没有知识卡")
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(cols, sep))
			fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(cols, sep))
			for _, n := range data.List {
				row := []ui.ColSpec{
					{Value: n.ID, Width: cols[0].Width},
					{Value: n.Name, Width: cols[1].Width},
					{Value: strconv.Itoa(len(n.Cards)), Width: cols[2].Width},
					{Value: n.CreateTime, Width: cols[3].Width},
				}
				fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n(page %d, %d items, total %d)\n", int(data.PageNum), len(data.List), int(data.Total))
			return nil
		},
	}
	cmd.Flags().IntVar(&page, "page", 1, "页码")
	cmd.Flags().IntVar(&limit, "limit", 10, "每页数量")
	return cmd
}
