package scene

import (
	"fmt"

	"github.com/BoteAI/zhizai-cli/cmd/cards"
	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

const sep = "  "

var sceneCols = []ui.ColSpec{
	{Value: "ID", Width: 12},
	{Value: "Name", Width: 28},
	{Value: "Category", Width: 16},
	{Value: "Desc", Width: 40},
}

// NewSceneCmd returns the scene command tree.
func NewSceneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scene",
		Short: "查询场景与知识卡",
		Example: `  zhizai scene list
  zhizai scene builtins
  zhizai scene cards
  zhizai cards`,
		RunE: func(c *cobra.Command, args []string) error {
			return c.Help()
		},
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newBuiltinsCmd())
	cmd.AddCommand(cards.NewCardsCmd())
	return cmd
}

func newListCmd() *cobra.Command {
	var includeShared bool

	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "列出我的场景（含共享可选）",
		Example: `  zhizai scene list
  zhizai scene list --include-shared=false
  zhizai scene list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryMySceneList"))
			list, err := c.SceneListMy(includeShared)
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), list)
			}
			printSceneTable(cmd, list, false)
			return nil
		},
	}

	cmd.Flags().BoolVar(&includeShared, "include-shared", true, "是否包含共享场景")
	return cmd
}

func newBuiltinsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "builtins",
		Args:  cobra.NoArgs,
		Short: "列出内置/平台场景（按分类展开）",
		Example: `  zhizai scene builtins
  zhizai scene builtins -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryInnerSceneList"))
			list, err := c.SceneListInner()
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), list)
			}
			printSceneTable(cmd, list, true)
			return nil
		},
	}
}

func printSceneTable(cmd *cobra.Command, list []client.Scene, withCategory bool) {
	cols := sceneCols
	if !withCategory {
		cols = []ui.ColSpec{
			{Value: "ID", Width: 12},
			{Value: "Name", Width: 28},
			{Value: "Desc", Width: 48},
		}
	}
	fmt.Fprint(cmd.OutOrStdout(), ui.PrintHeader(cols, sep))
	fmt.Fprint(cmd.OutOrStdout(), ui.DividerLine(cols, sep))
	for _, s := range list {
		var row []ui.ColSpec
		if withCategory {
			cat := s.CategoryName
			if cat == "" {
				cat = s.CategoryID
			}
			row = []ui.ColSpec{
				{Value: s.ID, Width: cols[0].Width},
				{Value: s.SceneName, Width: cols[1].Width},
				{Value: cat, Width: cols[2].Width},
				{Value: s.SceneDesc, Width: cols[3].Width},
			}
		} else {
			row = []ui.ColSpec{
				{Value: s.ID, Width: cols[0].Width},
				{Value: s.SceneName, Width: cols[1].Width},
				{Value: s.SceneDesc, Width: cols[2].Width},
			}
		}
		fmt.Fprint(cmd.OutOrStdout(), ui.PrintRow(row, sep))
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n(%d scenes)\n", len(list))
}
