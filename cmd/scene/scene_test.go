package scene

import (
	"testing"

	"github.com/BoteAI/zhizai-cli/cmd/cards"
)

func TestSceneCommandTree(t *testing.T) {
	root := NewSceneCmd()
	if root.Use != "scene" {
		t.Fatalf("use=%q", root.Use)
	}
	want := map[string]bool{"list": false, "builtins": false, "cards": false}
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

	if cards.NewCardsCmd().Name() != "cards" {
		t.Fatal("cards command name")
	}
}
