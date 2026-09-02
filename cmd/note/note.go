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
  zhizai note get <id> --field title
  zhizai note status <id>`,
	}

	root.AddCommand(newGetCmd())
	root.AddCommand(newStubCmd("create", "创建笔记"))
	root.AddCommand(newStubCmd("update", "更新笔记"))
	root.AddCommand(newStubCmd("delete", "删除笔记"))
	root.AddCommand(newStubCmd("status", "查询笔记处理进度"))

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
			note, err := client.New().NoteGet(args[0])
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

func newStubCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(c *cobra.Command, args []string) error {
			return output.NotImplemented("note " + use)
		},
	}
}
