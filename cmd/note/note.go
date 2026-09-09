package note

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/BoteAI/zhizai-cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewNoteCmd returns the note command tree.
func NewNoteCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "note",
		Short: "查看与管理单条笔记",
		Example: `  zhizai note get <id>
  zhizai note create --title 周报 --content "本周进展..."
  zhizai note update <id> --title 新标题
  zhizai note delete <id> --yes
  zhizai note status <id>`,
	}

	root.AddCommand(newGetCmd())
	root.AddCommand(newCreateCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newDeleteCmd())
	root.AddCommand(newStatusCmd())

	return root
}

func newGetCmd() *cobra.Command {
	var field string

	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "查看笔记详情",
		Args:  cobra.ExactArgs(1),
		Example: `  zhizai note get 30480
  zhizai note get 30480 --field summary
  zhizai note get 30480 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/querySingleNoteDetail"))
			note, err := c.NoteGet(args[0])
			if err != nil {
				return err
			}

			if field != "" {
				return printField(cmd, note, field)
			}

			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), note)
			}

			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("ID", note.ID))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Title", note.Title))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Type", client.NoteTypeLabel(note.NoteType)))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("State", note.NoteState))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Created", note.CreateTime))
			if note.SceneName != "" {
				fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Scene", note.SceneName))
			}
			if note.Abstract != "" {
				fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Abstract", ui.Truncate(note.Abstract, 160)))
			}
			if note.Summary != "" {
				fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Summary", ui.Truncate(note.Summary, 200)))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&field, "field", "", "只输出单个字段: id/title/type/state/summary/abstract/content/create_time/scene_name")
	return cmd
}

func newCreateCmd() *cobra.Command {
	var noteType string
	var title string
	var content string
	var sceneID string

	cmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.NoArgs,
		Short: "创建文字笔记",
		Long:  "当前仅支持 noteType=text。录音/文档/图片/链接需等待 file upload 能力。",
		Example: `  zhizai note create --title 周报 --content "本周完成..."
  zhizai note create --content "随手记一段文字" -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noteType == "" {
				noteType = "text"
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/createNote"))
			params := client.NoteCreateParams{
				NoteType: noteType,
				TextContent: &client.NoteTextContent{
					Title:   title,
					Content: content,
				},
			}
			if strings.TrimSpace(sceneID) != "" {
				params.SceneID = sceneID
			}
			note, err := c.NoteCreate(params)
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), note)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已创建笔记 %s\n", note.ID)
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Title", note.Title))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Type", client.NoteTypeLabel(note.NoteType)))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("State", note.NoteState))
			return nil
		},
	}

	cmd.Flags().StringVar(&noteType, "type", "text", "笔记类型（当前仅支持 text）")
	cmd.Flags().StringVar(&title, "title", "", "标题")
	cmd.Flags().StringVar(&content, "content", "", "正文")
	cmd.Flags().StringVar(&sceneID, "scene-id", "", "场景 ID（可选，用于按场景总结）")
	return cmd
}

func newUpdateCmd() *cobra.Command {
	var title string
	var abstract string
	var summary string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Args:  cobra.ExactArgs(1),
		Short: "更新笔记标题/摘要/总结",
		Example: `  zhizai note update 31560 --title 新标题
  zhizai note update 31560 --abstract 短摘要 --summary 长总结 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/updateNoteInfo"))
			if err := c.NoteUpdate(client.NoteUpdateParams{
				NoteID:          args[0],
				Title:           title,
				AbstractContent: abstract,
				Summary:         summary,
			}); err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]string{
					"noteId":   args[0],
					"status":   "updated",
					"title":    title,
					"abstract": abstract,
					"summary":  summary,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已更新笔记 %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "新标题")
	cmd.Flags().StringVar(&abstract, "abstract", "", "新摘要（对应 abstractContent）")
	cmd.Flags().StringVar(&summary, "summary", "", "新总结")
	return cmd
}

func newDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Args:  cobra.ExactArgs(1),
		Short: "删除笔记（需 --yes）",
		Example: `  zhizai note delete 31559 --yes
  zhizai note delete 31559 --yes -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("删除为不可逆操作，请加 --yes 确认")
			}
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/deleteNote"))
			if err := c.NoteDelete(args[0]); err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]string{
					"noteId": args[0],
					"status": "deleted",
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已删除笔记 %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "确认删除")
	return cmd
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <id>",
		Args:  cobra.ExactArgs(1),
		Short: "查询笔记处理进度",
		Example: `  zhizai note status 31560
  zhizai note status 31560 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New()
			fmt.Fprintf(cmd.ErrOrStderr(), "请求接口: %s\n", c.APIEndpoint("/note/queryNoteStatus"))
			st, err := c.NoteStatus(args[0])
			if err != nil {
				return err
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]string{
					"noteId":    args[0],
					"noteState": st.NoteState,
				})
			}
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("NoteID", args[0]))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("State", st.NoteState))
			fmt.Fprint(cmd.OutOrStdout(), ui.FormatFieldPair("Label", client.NoteStateLabel(st.NoteState)))
			return nil
		},
	}
}

func printField(cmd *cobra.Command, note *client.Note, field string) error {
	field = strings.ToLower(strings.TrimSpace(field))
	var value string
	switch field {
	case "id":
		value = note.ID
	case "title":
		value = note.Title
	case "type", "note_type":
		value = note.NoteType
	case "state", "note_state":
		value = note.NoteState
	case "summary":
		value = note.Summary
	case "abstract":
		value = note.Abstract
	case "create_time", "created":
		value = note.CreateTime
	case "scene_name", "scene":
		value = note.SceneName
	case "scene_id":
		value = note.SceneID
	case "content":
		if len(note.Content) == 0 {
			value = ""
		} else {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, note.Content, "", "  "); err == nil {
				value = pretty.String()
			} else {
				value = string(note.Content)
			}
		}
	default:
		return fmt.Errorf("不支持的字段 %q；可用: id, title, type, state, summary, abstract, content, create_time, scene_name", field)
	}

	if output.Format() == "json" {
		return output.WriteSuccessJSON(cmd.OutOrStdout(), map[string]string{field: value})
	}
	fmt.Fprintln(cmd.OutOrStdout(), value)
	return nil
}
