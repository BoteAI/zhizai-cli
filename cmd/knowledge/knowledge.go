package knowledge

import (
	"fmt"
	"strconv"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

const sep = "  "

var listCols = []ui.ColSpec{
	{Value: "ID", Width: 18},
	{Value: "Name", Width: 28},
	{Value: "Attr", Width: 10},
	{Value: "Notes", Width: 8},
	{Value: "Created", Width: 19},
}

var noteCols = []ui.ColSpec{
	{Value: "NoteID", Width: 18},
	{Value: "Title", Width: 32},
	{Value: "Type", Width: 10},
	{Value: "Updated", Width: 19},
}

// NewKnowledgeCmd returns the knowledge command tree.
func NewKnowledgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "管理笔记集",
		Long:  "查询我创建的 / 收到的笔记集，以及集内笔记详情。",
		Example: `  zhizai knowledge list
  zhizai knowledge list --received
  zhizai knowledge get <id>
  zhizai knowledge notes <id>`,
		RunE: func(c *cobra.Command, args []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newNotesCmd())
	cmd.AddCommand(newCreateCmd())
	return cmd
}

func newListCmd() *cobra.Command {
	var received bool
	var limit int
	var page int
	var query string

	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "列出笔记集",
		Example: `  zhizai knowledge list
  zhizai knowledge list --received
  zhizai knowledge list --limit 20 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			if page <= 0 {
				page = 1
			}
			if limit <= 0 {
				limit = 10
			}

			var (
				data *client.KnowledgePage
				err  error
			)
			if received {
				fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteKnowledgeEmpower"))
				data, err = c.KnowledgeReceived(client.KnowledgeReceivedParams{
					QryContent: query,
					PageNum:    page,
					PageSize:   limit,
				})
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteKnowledge"))
				data, err = c.KnowledgeList(client.KnowledgeListParams{
					QryType:    "myCreate",
					QryContent: query,
					PageNum:    page,
					PageSize:   limit,
				})
			}
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			printListHeader(cmd)
			for _, item := range data.List {
				printListRow(cmd, item)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n(page %d, %d items, total %d)\n", int(data.PageNum), len(data.List), int(data.Total))
			return nil
		},
	}

	cmd.Flags().BoolVar(&received, "received", false, "列出我收到的笔记集（默认我创建的）")
	cmd.Flags().IntVar(&limit, "limit", 10, "每页数量")
	cmd.Flags().IntVar(&page, "page", 1, "页码")
	cmd.Flags().StringVar(&query, "query", "", "搜索关键字")
	return cmd
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Args:  cobra.ExactArgs(1),
		Short: "查看笔记集详情",
		Example: `  zhizai knowledge get 10101
  zhizai knowledge get 10101 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteKnowledgeDetail"))
			detail, err := c.KnowledgeGet(client.KnowledgeDetailParams{
				KnowledgeID:        args[0],
				PageNum:            1,
				PageSize:           1,
				NeedSummaryContent: false,
			})
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), detail)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", detail.KnowledgeID)
			fmt.Fprintf(cmd.OutOrStdout(), "Name:    %s\n", detail.KnowledgeName)
			fmt.Fprintf(cmd.OutOrStdout(), "Detail:  %s\n", detail.KnowledgeDetail)
			fmt.Fprintf(cmd.OutOrStdout(), "Attr:    %s\n", detail.KnowledgeAttribute)
			fmt.Fprintf(cmd.OutOrStdout(), "Type:    %s\n", detail.KnowledgeType)
			fmt.Fprintf(cmd.OutOrStdout(), "Notes:   %d\n", int(detail.NoteTotal))
			fmt.Fprintf(cmd.OutOrStdout(), "Files:   %d\n", int(detail.FileTotal))
			fmt.Fprintf(cmd.OutOrStdout(), "Owner:   %s\n", detail.UserName)
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", detail.CreateTime)
			fmt.Fprintf(cmd.OutOrStdout(), "Editable:%v\n", detail.Editable)
			return nil
		},
	}
}

func newNotesCmd() *cobra.Command {
	var limit int
	var page int

	cmd := &cobra.Command{
		Use:   "notes <id>",
		Args:  cobra.ExactArgs(1),
		Short: "列出笔记集内的笔记",
		Example: `  zhizai knowledge notes 10101
  zhizai knowledge notes 10101 --limit 20 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if page <= 0 {
				page = 1
			}
			if limit <= 0 {
				limit = 10
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteKnowledgeDetail"))
			detail, err := c.KnowledgeGet(client.KnowledgeDetailParams{
				KnowledgeID:        args[0],
				PageNum:            page,
				PageSize:           limit,
				NeedSummaryContent: true,
			})
			if err != nil {
				return err
			}
			notes := detail.PageInfo
			if notes == nil {
				notes = &client.KnowledgeNotes{}
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]interface{}{
					"knowledgeId":   detail.KnowledgeID,
					"knowledgeName": detail.KnowledgeName,
					"pageInfo":      notes,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Knowledge: %s (%s)\n\n", detail.KnowledgeName, detail.KnowledgeID)
			printNoteHeader(cmd)
			for _, n := range notes.List {
				printNoteRow(cmd, n)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n(page %d, %d items, total %d)\n", int(notes.PageNum), len(notes.List), int(notes.Total))
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "每页数量")
	cmd.Flags().IntVar(&page, "page", 1, "页码")
	return cmd
}

func newCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.NoArgs,
		Short: "新建笔记集（接口待开放）",
		Example: `  zhizai knowledge create --name 项目周报`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name 不能为空")
			}
			_ = name
			return output.NotImplemented("knowledge create（后端 /note/createKnowledge 待补齐）")
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "笔记集名称")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func printListHeader(cmd *cobra.Command) {
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(listCols, sep))
	fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(listCols, sep))
}

func printListRow(cmd *cobra.Command, item client.Knowledge) {
	row := []ui.ColSpec{
		{Value: item.KnowledgeID, Width: listCols[0].Width},
		{Value: item.KnowledgeName, Width: listCols[1].Width},
		{Value: item.KnowledgeAttribute, Width: listCols[2].Width},
		{Value: strconv.Itoa(int(item.ContentTotal)), Width: listCols[3].Width},
		{Value: item.CreateTime, Width: listCols[4].Width},
	}
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
}

func printNoteHeader(cmd *cobra.Command) {
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(noteCols, sep))
	fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(noteCols, sep))
}

func printNoteRow(cmd *cobra.Command, n client.KnowledgeNoteItem) {
	row := []ui.ColSpec{
		{Value: n.NoteID, Width: noteCols[0].Width},
		{Value: n.Name, Width: noteCols[1].Width},
		{Value: client.NoteTypeLabel(n.NoteType), Width: noteCols[2].Width},
		{Value: n.NoteUpdateTime, Width: noteCols[3].Width},
	}
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
}
