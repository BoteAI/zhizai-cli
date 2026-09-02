package scene

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewSceneCmd returns the scene command tree.
func NewSceneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scene",
		Short: "查询场景与知识卡",
		Example: `  zhizai scene list`,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("scene")
		},
	}
}
