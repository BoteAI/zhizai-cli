package file

import (
	"fmt"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewFileCmd returns the file command tree.
func NewFileCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "file",
		Short: "上传与下载文件",
		Example: `  zhizai file upload ./meeting.mp3
  zhizai file upload ./a.jpg --compress -o json
  zhizai file get <fileId> --out ./out.bin`,
	}

	root.AddCommand(newUploadCmd())
	root.AddCommand(newGetCmd())
	return root
}

func newUploadCmd() *cobra.Command {
	var compress bool

	cmd := &cobra.Command{
		Use:   "upload <path>",
		Short: "上传本地文件，返回 fileId（不创建笔记）",
		Args:  cobra.ExactArgs(1),
		Example: `  zhizai file upload ./meeting.mp3
  zhizai file upload ./shot.png --compress -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := strings.TrimSpace(args[0])
			if path == "" {
				return fmt.Errorf("请指定要上传的文件路径")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/file/uploadSingleFile"))
			result, err := c.UploadFile(path, compress)
			if err != nil {
				return err
			}
			data := map[string]string{
				"fileId":   result.FileID,
				"fileName": result.FileName,
				"appId":    result.AppID,
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("FileID", result.FileID))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("FileName", result.FileName))
			if result.AppID != "" {
				fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("AppID", result.AppID))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&compress, "compress", false, "压缩后存储（音频不建议开启）")
	return cmd
}

func newGetCmd() *cobra.Command {
	var outPath string

	cmd := &cobra.Command{
		Use:   "get <fileId>",
		Short: "按 fileId 下载文件到本地",
		Args:  cobra.ExactArgs(1),
		Example: `  zhizai file get 1357179569765883904 --out ./out.bin
  zhizai file get 1357179569765883904 --out ./out.bin -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fileID := strings.TrimSpace(args[0])
			if fileID == "" {
				return fmt.Errorf("请指定 fileId")
			}
			outPath = strings.TrimSpace(outPath)
			if outPath == "" {
				return fmt.Errorf("请使用 --out 指定下载保存路径（-o 仅用于输出格式）")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/file/getFile/"+fileID))
			if err := c.DownloadFile(fileID, outPath); err != nil {
				return err
			}
			data := map[string]string{
				"fileId": fileID,
				"out":    outPath,
				"status": "downloaded",
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已下载到 %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&outPath, "out", "", "下载保存路径（必填）")
	return cmd
}
