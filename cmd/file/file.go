package file

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewFileCmd returns the file command tree.
func NewFileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "file",
		Short: "上传文件",
		Example: `  zhizai file upload <path>`,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("file")
		},
	}
}
