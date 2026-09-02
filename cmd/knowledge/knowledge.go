package knowledge

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewKnowledgeCmd returns the knowledge command tree.
func NewKnowledgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "knowledge",
		Short: "管理笔记集",
		Example: `  zhizai knowledge list`,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("knowledge")
		},
	}
}
