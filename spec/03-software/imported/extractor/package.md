# Package: extractor

<!-- reqmd: generated -->

## reqmd-code/internal/extractor/trace.go
### IMP-extractor-idShapeRe-3d989b4: idShapeRe
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/extractor/trace.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: var
trace: []
```

idShapeRe separates requirement IDs from symbol references. It is
deliberately looser than the schema's trace pattern, because an
ID-shaped token must reach `reqmd check` as a ref (where the canonical
pattern is enforced) instead of being mistaken for a missing symbol:

   - authored IDs: an uppercase first segment, at least one hyphen, and
     uppercase/digit segments (SW-PAR-001, REQ-MATH-ADD, REQ-1,
     system/REQ-1~2). The numeric-tail check lives in the schema, not here.
   - generated source-proxy IDs: prefix, package, symbol, and a lowercase
     7-hex hash (IMP-clang-Parse-479c6b0). At least two middle segments are
     required so a typo such as SW-imp-0000001 does not pass as generated.

A token matching either branch is passed through as an Item.Ref and
validated by `reqmd check`; everything else is a symbol reference resolved
by reqmd-code.

### IMP-extractor-traceRe-a46dbc7: traceRe
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/extractor/trace.go
x-reqmd.source-line: 34
x-reqmd.symbol-kind: var
trace: []
```

traceRe matches an explicit `reqmd:trace` directive line. One target per
line is canonical; a comma-separated list on one line is also accepted.

At most one leading space is tolerated: doc-comment code examples are
indented by a tab or four spaces, and must not become real traces.

### IMP-extractor-heuristicRe-74faede: heuristicRe
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/extractor/trace.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: var
trace: []
```

heuristicRe matches bare requirement IDs in free-form text. Used only
when the CLI flag --heuristic-traces is set. Requires at least 3 digits
in the trailing number to avoid false positives like UTF-8, HTTP-2.

Pattern: one uppercase-letter run, then 1+ "-<uppercase/digits>" segments,
then a "-<digits>" segment with 3+ digits, optionally followed by a
"~<digits>" version pin. This matches the multi-segment shape of real
requirement IDs (e.g. REQ-AUTH-001, SYS-LOG-002) without swallowing
short-numbered prose tokens like UTF-8 or HTTP-2.

### IMP-extractor-TraceRefs-6cc5980: TraceRefs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/extractor/trace.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: type
trace: []
```

TraceRefs is the classified result of scanning one item's doc text.

### IMP-extractor-ExtractTraces-d9a9734: ExtractTraces
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/extractor/trace.go
x-reqmd.source-line: 77
x-reqmd.symbol-kind: function
trace: []
```

ExtractTraces classifies the `reqmd:trace` directives in the given source
text. The text is expected to be the BindDoc-cleaned comment/docstring of
a symbol, but the function is language-agnostic.

Returned slices are stable, deduplicated, and preserve first-seen order.
A token found as a requirement ID never appears in SymbolRefs, and a
heuristic match never repeats an explicit requirement ID.

useHeuristic enables the broader bare-ID scan (driven by the CLI flag
`--heuristic-traces`).

## reqmd-code/internal/extractor/trace\_test.go
### IMP-extractor-TestExtractTraces-250f30c: TestExtractTraces
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/extractor/trace_test.go
x-reqmd.source-line: 8
x-reqmd.symbol-kind: function
trace:
  - IMP-extractor-ExtractTraces-d9a9734
```

