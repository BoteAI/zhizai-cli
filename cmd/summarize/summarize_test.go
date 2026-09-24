package summarize

import (
	"strings"
	"testing"
)

func TestSummarizeRequiresContent(t *testing.T) {
	cmd := NewSummarizeCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "content") {
		t.Fatalf("err=%v", err)
	}
}
