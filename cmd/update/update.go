package update

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewUpdateCmd returns the update command.
func NewUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "升级 CLI 并同步 Skill",
		Example: `  zhizai update
  zhizai update --check`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("update")
		},
	}
}
