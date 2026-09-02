package cmd

import (
	"fmt"

	"github.com/BoteAI/zhizai-cli/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Args:    cobra.NoArgs,
		Aliases: []string{"v"},
		Short:   "显示版本信息",
		Example: `  zhizai version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "zhizai version %s\n", version.Version)
			return nil
		},
	}
	return cmd
}
