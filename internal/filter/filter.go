// Package filter provides generic attribute-based filtering of requirements
// using expr-lang/expr as the evaluation engine.
//
// A filter expression is compiled once at startup into bytecode, then
// evaluated against each requirement's attribute map. The evaluation
// context exposes every attribute from the req's attr block plus the
// built-in variables `id` and `title` (and the RFC built-in attributes
// `status`, `disposition`, `trace`, `version`, which resolve to nil when
// a requirement does not declare them).
//
// Compile-time validation rejects expressions that reference attribute
// names not declared in any document's schema.yaml (a global typo check),
// and reports expr syntax errors once before any requirement is evaluated.
package filter

import (
	"fmt"
	"sort"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/expr-lang/expr/vm"

	"reqmd/internal/model"
)

// builtInVars are always available in the filter scope regardless of
// schema declarations. `id` and `title` live on model.Requirement, not in
// Attrs; the rest are the RFC §2.3 built-in attributes (status,
// disposition, trace, version). They are always accepted by the typo
// check even when no schema in the tree declares them; at evaluation time
// a requirement lacking the attribute sees nil (see AllowUndefinedVariables).
var builtInVars = map[string]struct{}{
	"id":          {},
	"title":       {},
	"status":      {},
	"disposition": {},
	"trace":       {},
	"version":     {},
}

// exprBuiltins is the set of expr built-in function names that are NOT
// variable references. We exclude them from attribute validation so that
// `id startsWith "SYS-"` or `len(variant) > 0` don't trip the typo check.
var exprBuiltins = map[string]struct{}{
	"len":        {},
	"all":        {},
	"any":        {},
	"none":       {},
	"one":        {},
	"filter":     {},
	"map":        {},
	"count":      {},
	"find":       {},
	"contains":   {},
	"startsWith": {},
	"endsWith":   {},
	"strings":    {},
}

// Filter is a compiled attribute filter expression. It is safe for
// concurrent use after construction — the underlying vm.Program is
// read-only.
type Filter struct {
	program *vm.Program
	expr    string
	refs    []string // validated variable names referenced in the expression
}

// Compile parses, validates, and compiles a filter expression.
//
// validAttrs is the union of all schema property names (plus built-in
// variable names) across every document in the tree. Referencing a name
// not in this set is a compile-time ERROR — this catches typos and
// references to attributes that no document declares.
//
// The expression is compiled with AllowUndefinedVariables so that a
// requirement whose document does not declare a referenced attribute
// evaluates that variable to nil (the req is excluded unless the
// expression handles absence, e.g. `variant == null`).
func Compile(exprStr string, validAttrs map[string]struct{}) (*Filter, error) {
	if exprStr == "" {
		return nil, fmt.Errorf("filter expression is empty")
	}

	// Phase 1: parse the AST and collect referenced variable names.
	tree, err := parser.Parse(exprStr)
	if err != nil {
		return nil, fmt.Errorf("filter syntax error: %w", err)
	}

	v := &refCollector{}
	ast.Walk(&tree.Node, v)

	// Phase 2: validate that every referenced name is either a known
	// attribute, a built-in variable, or an expr built-in function.
	for _, name := range v.refs {
		if _, isBuiltin := exprBuiltins[name]; isBuiltin {
			continue
		}
		if _, isVar := builtInVars[name]; isVar {
			continue
		}
		if _, isValid := validAttrs[name]; isValid {
			continue
		}
		return nil, fmt.Errorf("filter references unknown attribute %q: not declared in any schema.yaml", name)
	}

	// Phase 3: compile to bytecode. AllowUndefinedVariables lets
	// per-document attribute absence evaluate to nil instead of
	// erroring at runtime.
	program, err := expr.Compile(exprStr, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("compiling filter: %w", err)
	}

	// Dedup refs for the final list.
	seen := make(map[string]struct{}, len(v.refs))
	refs := make([]string, 0, len(v.refs))
	for _, r := range v.refs {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		refs = append(refs, r)
	}
	sort.Strings(refs)

	return &Filter{
		program: program,
		expr:    exprStr,
		refs:    refs,
	}, nil
}

// Expr returns the original expression string.
func (f *Filter) Expr() string { return f.expr }

// Refs returns the sorted, deduplicated list of variable names
// referenced in the expression (excluding expr built-in functions).
func (f *Filter) Refs() []string { return f.refs }

// Match evaluates the compiled expression against a single requirement.
// Returns (true, nil) if the requirement matches, (false, nil) if not,
// and (false, err) if evaluation fails (e.g. type mismatch).
func (f *Filter) Match(req model.Requirement) (bool, error) {
	env := buildEnv(req)
	result, err := expr.Run(f.program, env)
	if err != nil {
		return false, fmt.Errorf("filter %q on %s: %w", f.expr, req.ID, err)
	}
	b, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("filter %q on %s: expression did not evaluate to bool (got %T)", f.expr, req.ID, result)
	}
	return b, nil
}

// FilterDocs returns a new []model.Document containing only matching
// requirements per document. Documents with all requirements filtered
// out are kept (with an empty Requirements slice) so document-level
// structure is preserved for list/stats/export output.
func (f *Filter) FilterDocs(docs []model.Document) ([]model.Document, error) {
	out := make([]model.Document, len(docs))
	for i, doc := range docs {
		filtered := doc
		reqs := make([]model.Requirement, 0, len(doc.Requirements))
		for _, req := range doc.Requirements {
			match, err := f.Match(req)
			if err != nil {
				return nil, err
			}
			if match {
				reqs = append(reqs, req)
			}
		}
		filtered.Requirements = reqs
		out[i] = filtered
	}
	return out, nil
}

// MatchingIDs returns a set of requirement IDs that match the filter.
// Used by check/serve for filter-aware graph checks — the graph is
// built from the full tree, but only matching reqs are checked/reported.
func (f *Filter) MatchingIDs(docs []model.Document) (map[string]struct{}, error) {
	ids := make(map[string]struct{})
	for _, doc := range docs {
		for _, req := range doc.Requirements {
			match, err := f.Match(req)
			if err != nil {
				return nil, err
			}
			if match {
				ids[req.ID] = struct{}{}
			}
		}
	}
	return ids, nil
}

// buildEnv constructs the expr evaluation environment for a requirement:
// the req's Attrs map plus `id` and `title` injected as top-level vars.
// We copy into a new map to avoid mutating the original Attrs.
func buildEnv(req model.Requirement) map[string]any {
	env := make(map[string]any, len(req.Attrs)+2)
	for k, v := range req.Attrs {
		env[k] = v
	}
	env["id"] = req.ID
	env["title"] = req.Title
	return env
}

// BuildValidAttrs computes the union of all schema property names across
// every document plus the built-in variable names. The result is used by
// Compile for attribute validation.
func BuildValidAttrs(docs []model.Document) map[string]struct{} {
	valid := make(map[string]struct{}, 32)
	for _, doc := range docs {
		for _, prop := range doc.Properties {
			valid[prop] = struct{}{}
		}
	}
	for name := range builtInVars {
		valid[name] = struct{}{}
	}
	return valid
}

// CompileForDocs compiles an expression against the attribute union of the
// given documents. It is the shared entry point for every CLI command that
// accepts a --filter flag: Compile + BuildValidAttrs in one call. Returns
// (nil, nil) when expr is empty (no filter).
func CompileForDocs(docs []model.Document, expr string) (*Filter, error) {
	if expr == "" {
		return nil, nil
	}
	return Compile(expr, BuildValidAttrs(docs))
}

// refCollector implements ast.Visitor and collects all IdentifierNode
// values from the expression AST. These are cross-referenced against
// validAttrs to catch typos.
type refCollector struct {
	refs []string
}

func (v *refCollector) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		v.refs = append(v.refs, n.Value)
	}
}
