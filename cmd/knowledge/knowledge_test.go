package knowledge

import (
	"testing"
)

func TestKnowledgeCommandTree(t *testing.T) {
	root := NewKnowledgeCmd()
	if root.Use != "knowledge" {
		t.Fatalf("use=%q", root.Use)
	}
	aliases := root.Aliases
	if len(aliases) != 1 || aliases[0] != "kb" {
		t.Fatalf("aliases=%v", aliases)
	}

	want := map[string]bool{
		"list": false, "shared": false, "get": false, "notes": false,
		"create": false, "catalog": false, "mkdir": false, "add": false, "rmdir": false,
	}
	for _, c := range root.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, ok := range want {
		if !ok {
			t.Fatalf("missing subcommand %s", name)
		}
	}
}

func TestCollectNoteIDs(t *testing.T) {
	got := collectNoteIDs("1", "2, 3,1")
	if len(got) != 3 || got[0] != "1" || got[1] != "2" || got[2] != "3" {
		t.Fatalf("got=%v", got)
	}
	if len(collectNoteIDs("", "")) != 0 {
		t.Fatal("expected empty")
	}
}
