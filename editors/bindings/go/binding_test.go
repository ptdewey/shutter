package tree_sitter_snapshot_test

import (
	"testing"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_snapshot "github.com/ptdewey/shutter/bindings/go"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_snapshot.Language())
	if language == nil {
		t.Errorf("Error loading Snapshot grammar")
	}
}
