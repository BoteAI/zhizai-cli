package ask

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAskEmptyQueryFails(t *testing.T) {
	cmd := NewAskCmd()
	cmd.SetArgs([]string{})
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "问题不能为空") {
		t.Fatalf("err=%v", err)
	}
}

func TestAskNoSessionWithoutChatIDFails(t *testing.T) {
	cmd := NewAskCmd()
	cmd.SetArgs([]string{"hello", "--no-session"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "请先创建问小智会话") {
		t.Fatalf("err=%v", err)
	}
}

func TestAskSessionEmptyQueryFails(t *testing.T) {
	root := NewAskCmd()
	root.SetArgs([]string{"session"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "问题不能为空") {
		t.Fatalf("err=%v", err)
	}
}

func TestAskChatRequiresIDs(t *testing.T) {
	root := NewAskCmd()
	root.SetArgs([]string{"chat", "追问"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "chat-id") {
		t.Fatalf("err=%v", err)
	}
}

func TestAskRefsRequiresDocIDs(t *testing.T) {
	root := NewAskCmd()
	root.SetArgs([]string{"refs", "--chat-id", "c", "--context-id", "x"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "doc-ids") {
		t.Fatalf("err=%v", err)
	}
}

func TestAskTemplateEmptyFails(t *testing.T) {
	root := NewAskCmd()
	root.SetArgs([]string{"template"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "模版指令不能为空") {
		t.Fatalf("err=%v", err)
	}
}

func TestResolveScopeNoteIDs(t *testing.T) {
	s := resolveScope(scopeFlags{note: []string{"a"}, noteIDs: "b, c"})
	if len(s.NoteIDList) != 3 {
		t.Fatalf("got %v", s.NoteIDList)
	}
}

// Ensure default ask remains a runnable cobra command with subcommands registered.
func TestAskSubcommandsRegistered(t *testing.T) {
	root := NewAskCmd()
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"session", "chat", "refs", "template"} {
		if !names[want] {
			t.Fatalf("missing subcommand %s", want)
		}
	}
	_ = cobra.Command{}
}
