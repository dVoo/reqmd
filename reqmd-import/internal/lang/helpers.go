package lang

import (
	"sort"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"reqmd-import/internal/model"
)

// CaptureContent returns the textual content of a captured node, or the
// empty string when the node is nil. It exists as a shared helper
// because every language plugin needs the same nil-guarded access to
// tree-sitter node text; previously each plugin reimplemented this
// 3-liner.
func CaptureContent(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}
	return node.Utf8Text(src)
}

// SortByStartLine sorts symbols in place by StartLine ascending.
// Hoisted because every plugin's Parse returns symbols in source
// order but wants a defensive re-sort as the last step.
func SortByStartLine(s []model.Symbol) {
	sort.Slice(s, func(i, j int) bool { return s[i].StartLine < s[j].StartLine })
}

// SymbolAt fills in the span and doc fields for a single symbol. The
// language parameter is the interface (not the concrete plugin type)
// so this helper is reusable across plugins: each plugin's only
// dependency on its own type was the call to BindDoc, which is part
// of the Language contract.
func SymbolAt(node *sitter.Node, file, pkg string, kind model.SymbolKind, name, receiver string, l Language, src []byte) model.Symbol {
	return model.Symbol{
		File:      file,
		Package:   pkg,
		Kind:      kind,
		Name:      name,
		Receiver:  receiver,
		Doc:       l.BindDoc(node, src),
		StartLine: int(node.StartPosition().Row) + 1,
		EndLine:   int(node.EndPosition().Row) + 1,
	}
}
