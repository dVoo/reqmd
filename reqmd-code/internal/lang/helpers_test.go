package lang_test

import (
	"testing"

	"github.com/dVoo/reqmd/reqmd-code/internal/lang"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

// TestModuleNameFromFile covers the helper every plugin uses to derive a
// default package name from the source path (one shared table instead of one
// per language).
func TestModuleNameFromFile(t *testing.T) {
	cases := []struct{ in, want string }{
		{"foo.py", "foo"},
		{"path/to/bar.c", "bar"},
		{"/abs/path/baz.rs", "baz"},
		{"nested/dir.with.dot/qux.go", "qux"},
		{`windows\dir\file.cpp`, "file"},
		{"", ""},
		{"noext", "noext"},
	}
	for _, c := range cases {
		if got := lang.ModuleNameFromFile(c.in); got != c.want {
			t.Errorf("ModuleNameFromFile(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestHasAncestorKind covers the shared ancestor walk: the nil guard, a
// matching ancestor, and one of several candidate kinds.
func TestHasAncestorKind(t *testing.T) {
	// func_literal > block > var_declaration is the shape every plugin filters on.
	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(sitter.NewLanguage(tree_sitter_go.Language())); err != nil {
		t.Fatalf("set language: %v", err)
	}

	src := []byte("package p\n\nvar f = func() {\n\tvar x int\n}\n")
	tree := parser.Parse(src, nil)
	if tree == nil {
		t.Fatal("parse failed")
	}
	defer tree.Close()

	// inner is the var_declaration on row 3 (inside the func literal);
	// the row-2 one wraps the literal itself.
	var inner *sitter.Node
	walk(tree.RootNode(), func(n *sitter.Node) {
		if n.Kind() == "var_declaration" && n.StartPosition().Row == 3 {
			inner = n
		}
	})
	if inner == nil {
		t.Fatal("var_declaration inside func literal not found")
	}
	if !lang.HasAncestorKind(inner, "function_declaration", "func_literal") {
		t.Error("expected func_literal ancestor to be detected")
	}
	if lang.HasAncestorKind(inner, "type_declaration") {
		t.Error("unexpected type_declaration ancestor")
	}
	if lang.HasAncestorKind(nil, "func_literal") {
		t.Error("nil node must not match")
	}
}

// walk visits every node depth-first.
func walk(n *sitter.Node, fn func(*sitter.Node)) {
	fn(n)
	for i := uint(0); i < n.ChildCount(); i++ {
		walk(n.Child(i), fn)
	}
}
