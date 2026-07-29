package model

// GeneratedReq is the writer-facing view of one requirement to be emitted
// into a .md file. The extractor (extract package) produces these from the
// language plugins' Symbol output plus any trace inference and ID stamping.
//
// The fields map 1:1 to the attr-block keys and body content that the
// downstream reqmd CLI consumes; the writer renders them in a stable
// order so re-runs produce byte-identical output (which is what makes the
// idempotent write-if-changed path work).
type GeneratedReq struct {
	// ID is the unique requirement identifier. The standard scheme is
	// "<id-prefix><Package>-<Symbol>-<7charhash>", e.g. "IMP-foo-FooBar-x7y3k2a".
	// The hash is computed over (package, symbol name) so IDs survive
	// line-number shifts (refactors that move code do not invalidate IDs).
	ID string

	// Title is the human-readable requirement title, typically the symbol's
	// name or "<Receiver>.<Name>" for methods. Empty means render the
	// heading without a title suffix.
	Title string

	// Status is the reqmd built-in status. The default for generated items
	// is "approved" so they count as upstream coverage providers.
	Status string

	// Body is the markdown body that follows the attr block. Typically the
	// symbol's doc comment, cleaned of comment markers.
	Body string

	// SourceFile is the source file path, rendered into the
	// x-reqmd.source-file attr key for provenance.
	SourceFile string

	// SourceLine is the 1-based line number in SourceFile, rendered into
	// x-reqmd.source-line. Zero means omit.
	SourceLine int

	// SymbolKind is the original SymbolKind value, rendered into
	// x-reqmd.symbol-kind. Empty means omit.
	SymbolKind string

	// ParentID links this requirement to a parent requirement as a
	// sub-requirement. For Go, methods inside a struct set ParentID to the
	// struct's requirement ID, which the writer renders as `###` instead of
	// `##`. Empty means top-level.
	ParentID string

	// Trace is the list of upstream requirement IDs this symbol implements
	// or is derived from. Rendered into the `trace` attr key as a YAML
	// list. Empty renders as `trace: []`.
	Trace []string

	// HeuristicTrace is the list of requirement IDs inferred from bare
	// occurrences in the source code (e.g. "See REQ-AUTH-001 for details")
	// when reqmd-import runs with --heuristic-traces. Rendered as a
	// separate x-reqmd.heuristic-trace attribute in the .md output to keep
	// the developer-asserted `trace:` list clean. Empty means omit the
	// attribute entirely.
	HeuristicTrace []string

	// Symbol is the original Symbol this requirement was derived from.
	// Not rendered into the .md; carried for downstream consumers and
	// for test fixtures.
	Symbol Symbol
}
