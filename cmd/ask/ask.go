package ask

import (
	"fmt"
	"strings"

	"github.com/BoteAI/zhizai-cli/internal/client"
	"github.com/BoteAI/zhizai-cli/internal/output"
	"github.com/spf13/cobra"
)

type scopeFlags struct {
	kb      []string
	dir     []string
	note    []string
	noteIDs string
}

type chatFlags struct {
	scopeFlags
	withReferences bool
	chatID         string
	contextID      string
	noSession      bool
}

// NewAskCmd returns the ask command tree (xiaozhi + template).
func NewAskCmd() *cobra.Command {
	var f chatFlags

	cmd := &cobra.Command{
		Use:   "ask [question]",
		Short: "问小智语义问答与动态模版",
		Long: `默认：自动创建会话后流式问答（可用 --chat-id/--context-id 追问）。
子命令：session / chat / refs / template。
--no-session 且未提供 --chat-id 时失败。`,
		Example: `  zhizai ask "找找关于支付的笔记" -o json
  zhizai ask "渠道下沉怎么安排的" --kb <knowledgeId> -o json
  zhizai ask "待办有哪些" --chat-id <id> --context-id <id> -o json
  zhizai ask session --query "找找关于支付的笔记" -o json
  zhizai ask chat "那收入增长是多少" --chat-id <id> --context-id <id> -o json
  zhizai ask refs --doc-ids 1,2 --context-id <id> -o json
  zhizai ask template "生成周报" --with-notes --start "..." --end "..." -o json`,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			return runDefaultAsk(cmd, query, f)
		},
	}

	bindScopeFlags(cmd, &f.scopeFlags)
	cmd.Flags().BoolVar(&f.withReferences, "with-references", false, "问答时请求参考文档")
	cmd.Flags().StringVar(&f.chatID, "chat-id", "", "已有会话 chatId（追问）")
	cmd.Flags().StringVar(&f.contextID, "context-id", "", "已有会话 contextId（追问）")
	cmd.Flags().BoolVar(&f.noSession, "no-session", false, "禁止自动建会话；无 --chat-id 时失败")

	cmd.AddCommand(newSessionCmd())
	cmd.AddCommand(newChatCmd())
	cmd.AddCommand(newRefsCmd())
	cmd.AddCommand(newTemplateCmd())
	return cmd
}

func bindScopeFlags(cmd *cobra.Command, f *scopeFlags) {
	cmd.Flags().StringArrayVar(&f.kb, "kb", nil, "限定笔记集 ID（可多次）")
	cmd.Flags().StringArrayVar(&f.dir, "dir", nil, "限定目录 ID（可多次；需同时带所属 --kb）")
	cmd.Flags().StringArrayVar(&f.note, "note", nil, "限定笔记 ID（可多次）")
	cmd.Flags().StringVar(&f.noteIDs, "note-ids", "", "限定笔记 ID（逗号分隔）")
}

func runDefaultAsk(cmd *cobra.Command, query string, f chatFlags) error {
	if query == "" {
		return fmt.Errorf("问题不能为空")
	}
	if f.noSession && strings.TrimSpace(f.chatID) == "" {
		return fmt.Errorf("请先创建问小智会话（--no-session 且未提供 --chat-id）")
	}

	c := client.New()
	scope := resolveScope(f.scopeFlags)

	chatID := strings.TrimSpace(f.chatID)
	contextID := strings.TrimSpace(f.contextID)

	if chatID == "" {
		sess, err := c.CreateXiaozhiSession(buildSessionParams(query, scope))
		if err != nil {
			return err
		}
		chatID = sess.ChatID
		contextID = sess.ContextID
	} else if contextID == "" {
		return fmt.Errorf("追问时必须同时提供 --context-id")
	}

	return runChatAndOutput(cmd, c, query, chatID, contextID, f.withReferences, scope)
}

func newSessionCmd() *cobra.Command {
	var query string
	var scope scopeFlags

	cmd := &cobra.Command{
		Use:   "session",
		Short: "仅创建问小智会话，返回 chatId/contextId",
		Example: `  zhizai ask session --query "找找关于支付的笔记" -o json
  zhizai ask session --query "待办有哪些" --kb <knowledgeId> -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query = strings.TrimSpace(query)
			if query == "" {
				return fmt.Errorf("问题不能为空（请传 --query）")
			}
			c := client.New()
			sess, err := c.CreateXiaozhiSession(buildSessionParams(query, resolveScope(scope)))
			if err != nil {
				return err
			}
			return writeSession(cmd, sess)
		},
	}
	cmd.Flags().StringVar(&query, "query", "", "问题（同时作为会话标题）")
	bindScopeFlags(cmd, &scope)
	return cmd
}

func newChatCmd() *cobra.Command {
	var f chatFlags
	cmd := &cobra.Command{
		Use:   "chat [question]",
		Short: "仅流式问答（不自动建会话）",
		Example: `  zhizai ask chat "那收入增长是多少" --chat-id <id> --context-id <id> -o json`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return fmt.Errorf("问题不能为空")
			}
			chatID := strings.TrimSpace(f.chatID)
			contextID := strings.TrimSpace(f.contextID)
			if chatID == "" {
				return fmt.Errorf("请提供 --chat-id（chat 子命令不会自动建会话）")
			}
			if contextID == "" {
				return fmt.Errorf("请提供 --context-id")
			}
			c := client.New()
			return runChatAndOutput(cmd, c, query, chatID, contextID, f.withReferences, resolveScope(f.scopeFlags))
		},
	}
	bindScopeFlags(cmd, &f.scopeFlags)
	cmd.Flags().BoolVar(&f.withReferences, "with-references", false, "问答时请求参考文档")
	cmd.Flags().StringVar(&f.chatID, "chat-id", "", "会话 chatId")
	cmd.Flags().StringVar(&f.contextID, "context-id", "", "会话 contextId")
	return cmd
}

func newRefsCmd() *cobra.Command {
	var chatID, contextID, docIDs string
	cmd := &cobra.Command{
		Use:   "refs",
		Short: "查询问小智参考文档详情",
		Example: `  zhizai ask refs --doc-ids 1,2,3 --context-id <id> -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := splitCSV(docIDs)
			if len(ids) == 0 {
				return fmt.Errorf("请提供 --doc-ids（来自问答 SSE references 事件）")
			}
			c := client.New()
			refs, err := c.XiaozhiRefs(ids)
			if err != nil {
				return err
			}
			if cid := strings.TrimSpace(contextID); cid != "" {
				_ = c.ClearXiaozhiCache(cid)
			}
			_ = chatID // kept for CLI compatibility with Excel flags
			data := map[string]any{
				"references": refs,
				"docIds":     ids,
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			if len(refs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no references)")
				return nil
			}
			for _, r := range refs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", r.NoteID, r.Name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&docIDs, "doc-ids", "", "引用 id，逗号分隔")
	cmd.Flags().StringVar(&chatID, "chat-id", "", "会话 chatId（可选）")
	cmd.Flags().StringVar(&contextID, "context-id", "", "会话 contextId（有则清缓存）")
	return cmd
}

func newTemplateCmd() *cobra.Command {
	var withNotes bool
	var start, end string
	cmd := &cobra.Command{
		Use:   "template [command]",
		Short: "查询动态成文模版（可选附带笔记列表）",
		Long:  "CLI 只返回模版与笔记 JSON，不调用 LLM 成文；由 Agent 填空。",
		Example: `  zhizai ask template "生成周报" -o json
  zhizai ask template "生成周报" --with-notes --start "2026-09-01 00:00:00" --end "2026-09-07 23:59:59" -o json`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.TrimSpace(strings.Join(args, " "))
			if command == "" {
				return fmt.Errorf("模版指令不能为空")
			}
			c := client.New()
			tpl, err := c.QueryTemplate(command)
			if err != nil {
				return err
			}
			// Retry with same command already done by API once; if empty output, caller may retry full sentence.
			data := map[string]any{
				"command":  command,
				"template": tpl,
			}
			if withNotes {
				notes, err := c.NoteList(client.NoteListParams{
					StartTime: strings.TrimSpace(start),
					EndTime:   strings.TrimSpace(end),
					PageNum:   1,
					PageSize:  100,
				})
				if err != nil {
					return err
				}
				data["notes"] = notes
			}
			if output.Format() == "json" {
				return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "input:\n%s\n\noutput:\n%s\n", tpl.Input, tpl.Output)
			if withNotes {
				if notes, ok := data["notes"].(*client.NoteListData); ok {
					fmt.Fprintf(cmd.OutOrStdout(), "\nnotes: %d\n", len(notes.List))
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&withNotes, "with-notes", false, "同时查询笔记列表供 Agent 填模版")
	cmd.Flags().StringVar(&start, "start", "", "笔记开始时间")
	cmd.Flags().StringVar(&end, "end", "", "笔记结束时间")
	return cmd
}

type resolvedScope struct {
	KnowledgeID     string
	KnowledgeIDList []string
	NoteID          string
	NoteIDList      []string
	DirectoryID     string
	DirectoryIDList []string
}

func resolveScope(f scopeFlags) resolvedScope {
	notes := append([]string{}, f.note...)
	notes = append(notes, splitCSV(f.noteIDs)...)
	return resolvedScope{
		KnowledgeIDList: clean(f.kb),
		DirectoryIDList: clean(f.dir),
		NoteIDList:      clean(notes),
	}
}

func clean(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	return clean(parts)
}

func buildSessionParams(query string, scope resolvedScope) client.XiaozhiSessionParams {
	p := client.XiaozhiSessionParams{Query: query}
	applyScopeToSession(&p, scope)
	return p
}

func applyScopeToSession(p *client.XiaozhiSessionParams, scope resolvedScope) {
	switch len(scope.KnowledgeIDList) {
	case 0:
	case 1:
		p.KnowledgeID = scope.KnowledgeIDList[0]
	default:
		p.KnowledgeIDList = scope.KnowledgeIDList
	}
	switch len(scope.NoteIDList) {
	case 0:
	case 1:
		p.NoteID = scope.NoteIDList[0]
	default:
		p.NoteIDList = scope.NoteIDList
	}
	switch len(scope.DirectoryIDList) {
	case 0:
	case 1:
		p.DirectoryID = scope.DirectoryIDList[0]
	default:
		p.DirectoryIDList = scope.DirectoryIDList
	}
}

func applyScopeToChat(p *client.XiaozhiChatParams, scope resolvedScope) {
	switch len(scope.KnowledgeIDList) {
	case 0:
	case 1:
		p.KnowledgeID = scope.KnowledgeIDList[0]
	default:
		p.KnowledgeIDList = scope.KnowledgeIDList
	}
	switch len(scope.NoteIDList) {
	case 0:
	case 1:
		p.NoteID = scope.NoteIDList[0]
	default:
		p.NoteIDList = scope.NoteIDList
	}
	switch len(scope.DirectoryIDList) {
	case 0:
	case 1:
		p.DirectoryID = scope.DirectoryIDList[0]
	default:
		p.DirectoryIDList = scope.DirectoryIDList
	}
}

func runChatAndOutput(cmd *cobra.Command, c *client.Client, query, chatID, contextID string, withRefs bool, scope resolvedScope) error {
	params := client.XiaozhiChatParams{
		Query:          query,
		ChatID:         chatID,
		ContextID:      contextID,
		WithReferences: withRefs,
	}
	applyScopeToChat(&params, scope)

	chat, err := c.XiaozhiChat(params)
	if err != nil {
		return err
	}

	var refs []client.XiaozhiReference
	if withRefs && len(chat.ReferenceIDs) > 0 {
		refs, err = c.XiaozhiRefs(chat.ReferenceIDs)
		if err != nil {
			return err
		}
	}
	_ = c.ClearXiaozhiCache(contextID)

	data := map[string]any{
		"chatId":        chatID,
		"contextId":     contextID,
		"query":         query,
		"answer":        chat.Text,
		"referenceIds":  chat.ReferenceIDs,
		"references":    refs,
	}
	if output.Format() == "json" {
		return output.WriteSuccessJSON(cmd.OutOrStdout(), data)
	}
	fmt.Fprintln(cmd.OutOrStdout(), chat.Text)
	if len(refs) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\n--- references ---")
		for _, r := range refs {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", r.NoteID, r.Name)
		}
	}
	return nil
}

func writeSession(cmd *cobra.Command, sess *client.XiaozhiSession) error {
	if output.Format() == "json" {
		return output.WriteSuccessJSON(cmd.OutOrStdout(), sess)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "chatId: %s\ncontextId: %s\nsessionTitle: %s\n",
		sess.ChatID, sess.ContextID, sess.SessionTitle)
	return nil
}
