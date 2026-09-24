package knowledge

import (
	"fmt"
	"strconv"
	"strings"

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

var catalogCols = []ui.ColSpec{
	{Value: "CatalogID", Width: 20},
	{Value: "Name", Width: 28},
	{Value: "Parent", Width: 20},
	{Value: "Notes", Width: 8},
	{Value: "Dirs", Width: 8},
}

// NewKnowledgeCmd returns the knowledge command tree (alias: kb).
func NewKnowledgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "knowledge",
		Aliases: []string{"kb"},
		Short:   "管理笔记集",
		Long:    "查询我创建的 / 收到的笔记集，以及目录与集内笔记。可用别名 kb（如 zhizai kb list）。",
		Example: `  zhizai knowledge list
  zhizai kb list --received
  zhizai kb shared
  zhizai kb get <id>
  zhizai kb create --name 项目周报
  zhizai kb catalog --kb <id>
  zhizai kb mkdir --kb <id> --name 周报
  zhizai kb add --kb <id> --note-id <noteId>
  zhizai kb rmdir <directoryId> --yes
  zhizai kb notes <id>`,
		RunE: func(c *cobra.Command, args []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSharedCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newNotesCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newCatalogCmd())
	cmd.AddCommand(newMkdirCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRmdirCmd())
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
  zhizai kb list --received
  zhizai kb shared
  zhizai knowledge list --limit 20 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, received, limit, page, query)
		},
	}

	cmd.Flags().BoolVar(&received, "received", false, "列出我收到的笔记集（默认我创建的；等价 kb shared）")
	cmd.Flags().IntVar(&limit, "limit", 10, "每页数量")
	cmd.Flags().IntVar(&page, "page", 1, "页码")
	cmd.Flags().StringVar(&query, "query", "", "搜索关键字")
	return cmd
}

func newSharedCmd() *cobra.Command {
	var limit int
	var page int
	var query string

	cmd := &cobra.Command{
		Use:   "shared",
		Args:  cobra.NoArgs,
		Short: "列出我收到的笔记集（等价 list --received）",
		Example: `  zhizai kb shared
  zhizai kb shared --limit 20 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, true, limit, page, query)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "每页数量")
	cmd.Flags().IntVar(&page, "page", 1, "页码")
	cmd.Flags().StringVar(&query, "query", "", "搜索关键字")
	return cmd
}

func runList(cmd *cobra.Command, received bool, limit, page int, query string) error {
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
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Args:  cobra.ExactArgs(1),
		Short: "查看笔记集详情",
		Example: `  zhizai knowledge get 10101
  zhizai kb get 10101 -o json`,
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
	var dir string

	cmd := &cobra.Command{
		Use:   "notes <id>",
		Args:  cobra.ExactArgs(1),
		Short: "列出笔记集内的笔记",
		Example: `  zhizai knowledge notes 10101
  zhizai kb notes 10101 --dir -1 --limit 20 -o json`,
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
				DirectoryID:        strings.TrimSpace(dir),
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
	cmd.Flags().StringVar(&dir, "dir", "", "目录 ID（空=整个笔记集；-1=根层）")
	return cmd
}

func newCreateCmd() *cobra.Command {
	var name string
	var detail string
	var attr string

	cmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.NoArgs,
		Short: "新建笔记集",
		Example: `  zhizai knowledge create --name 项目周报
  zhizai kb create --name 工作笔记集 --detail 周会纪要 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("--name 不能为空")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/createNoteKnowledge"))
			id, err := c.CreateKnowledge(client.CreateKnowledgeParams{
				KnowledgeName:      name,
				KnowledgeDetail:    strings.TrimSpace(detail),
				KnowledgeAttribute: strings.TrimSpace(attr),
			})
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]string{
					"knowledgeId":   id,
					"knowledgeName": name,
				})
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("KnowledgeID", id))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Name", name))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "笔记集名称")
	cmd.Flags().StringVar(&detail, "detail", "", "笔记集描述")
	cmd.Flags().StringVar(&attr, "attr", "", "属性（默认 private）")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newCatalogCmd() *cobra.Command {
	var kb string
	var parent string

	cmd := &cobra.Command{
		Use:   "catalog",
		Args:  cobra.NoArgs,
		Short: "列出笔记集目录",
		Example: `  zhizai kb catalog --kb 10101
  zhizai kb catalog --kb 10101 --parent -1 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kb = strings.TrimSpace(kb)
			if kb == "" {
				return fmt.Errorf("--kb 不能为空")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryKnowledgeCatalog"))
			list, err := c.KnowledgeCatalogList(client.KnowledgeCatalogParams{
				KnowledgeID: kb,
				CatalogID:   strings.TrimSpace(parent),
			})
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), list)
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(catalogCols, sep))
			fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(catalogCols, sep))
			for _, item := range list {
				row := []ui.ColSpec{
					{Value: item.CatalogID, Width: catalogCols[0].Width},
					{Value: item.CatalogName, Width: catalogCols[1].Width},
					{Value: item.ParentCatalogID, Width: catalogCols[2].Width},
					{Value: strconv.Itoa(int(item.DetailTotal)), Width: catalogCols[3].Width},
					{Value: strconv.Itoa(int(item.CatalogTotal)), Width: catalogCols[4].Width},
				}
				fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n(%d directories)\n", len(list))
			return nil
		},
	}

	cmd.Flags().StringVar(&kb, "kb", "", "笔记集 ID")
	cmd.Flags().StringVar(&parent, "parent", "", "父目录 ID（默认 -1 根层）")
	_ = cmd.MarkFlagRequired("kb")
	return cmd
}

func newMkdirCmd() *cobra.Command {
	var kb string
	var name string
	var parent string

	cmd := &cobra.Command{
		Use:   "mkdir",
		Args:  cobra.NoArgs,
		Short: "在笔记集下新建目录",
		Example: `  zhizai kb mkdir --kb 10101 --name 周报
  zhizai kb mkdir --kb 10101 --name 子目录 --parent <catalogId> -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kb = strings.TrimSpace(kb)
			name = strings.TrimSpace(name)
			if kb == "" {
				return fmt.Errorf("--kb 不能为空")
			}
			if name == "" {
				return fmt.Errorf("--name 不能为空")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/addKnowledgeCatalog"))
			cat, err := c.AddKnowledgeCatalog(client.AddKnowledgeCatalogParams{
				KnowledgeID:       kb,
				DirectoryName:     name,
				ParentDirectoryID: strings.TrimSpace(parent),
			})
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), cat)
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("CatalogID", cat.CatalogID))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Name", cat.CatalogName))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("KnowledgeID", cat.KnowledgeID))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Parent", cat.ParentCatalogID))
			return nil
		},
	}

	cmd.Flags().StringVar(&kb, "kb", "", "笔记集 ID")
	cmd.Flags().StringVar(&name, "name", "", "目录名称")
	cmd.Flags().StringVar(&parent, "parent", "", "父目录 ID（默认 -1 根层）")
	_ = cmd.MarkFlagRequired("kb")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newAddCmd() *cobra.Command {
	var kb string
	var noteID string
	var noteIDs string
	var dir string

	cmd := &cobra.Command{
		Use:   "add",
		Args:  cobra.NoArgs,
		Short: "将笔记加入或移动到笔记集目录",
		Example: `  zhizai kb add --kb 10101 --note-id 31023
  zhizai kb add --kb 10101 --note-ids 31023,31024 --dir <catalogId> -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			kb = strings.TrimSpace(kb)
			if kb == "" {
				return fmt.Errorf("--kb 不能为空")
			}
			ids := collectNoteIDs(noteID, noteIDs)
			if len(ids) == 0 {
				return fmt.Errorf("请指定 --note-id 或 --note-ids")
			}
			params := client.MoveNotesToKnowledgeParams{
				KnowledgeID: kb,
				DirectoryID: strings.TrimSpace(dir),
			}
			if len(ids) == 1 {
				params.NoteID = ids[0]
			} else {
				params.NoteIDList = ids
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/moveNotesToKnowledge"))
			if err := c.MoveNotesToKnowledge(params); err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]interface{}{
					"knowledgeId": kb,
					"noteIds":     ids,
					"directoryId": strings.TrimSpace(dir),
					"status":      "moved",
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已将 %d 条笔记加入笔记集 %s\n", len(ids), kb)
			return nil
		},
	}

	cmd.Flags().StringVar(&kb, "kb", "", "笔记集 ID")
	cmd.Flags().StringVar(&noteID, "note-id", "", "单个笔记 ID")
	cmd.Flags().StringVar(&noteIDs, "note-ids", "", "多个笔记 ID（逗号分隔）")
	cmd.Flags().StringVar(&dir, "dir", "", "目标目录 ID（默认根层）")
	_ = cmd.MarkFlagRequired("kb")
	return cmd
}

func newRmdirCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "rmdir <directoryId>",
		Args:  cobra.ExactArgs(1),
		Short: "删除笔记集目录（需 --yes；不删笔记正文）",
		Example: `  zhizai kb rmdir 30101 --yes
  zhizai kb rmdir 30101 --yes -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("删除目录会级联去掉子目录与笔记关联，请加 --yes 确认")
			}
			dirID := strings.TrimSpace(args[0])
			if dirID == "" {
				return fmt.Errorf("directoryId 不能为空")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/deleteKnowledgeCatalog"))
			if err := c.DeleteKnowledgeCatalog(dirID); err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]string{
					"directoryId": dirID,
					"status":      "deleted",
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已删除目录 %s\n", dirID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "确认删除")
	return cmd
}

func collectNoteIDs(single, csv string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		for _, part := range strings.Split(raw, ",") {
			id := strings.TrimSpace(part)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	add(single)
	add(csv)
	return out
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
