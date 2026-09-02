package msg

import (
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewMsgCmd returns the msg command tree.
func NewMsgCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "msg",
		Short: "发送消息与查询录音卡",
		Example: `  zhizai msg send`,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("msg")
		},
	}
}
