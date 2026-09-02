package team

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewTeamCmd returns the team command tree.
func NewTeamCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "team",
		Short: "管理团队与成员",
		Example: `  zhizai team list`,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("team")
		},
	}
}
