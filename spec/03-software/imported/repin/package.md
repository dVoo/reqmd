# Package: repin

<!-- reqmd: generated -->

## internal/repin/repin.go
### IMP-repin-Result-e706624: Result
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 31
x-reqmd.symbol-kind: type
trace: []
```

Result is the value Build returns: a sorted list of deltas plus a
per-file summary derived from the same list. The caller passes
Deltas to Apply or to one of the Format* functions.

The JSON field names form the stable contract for the
`reqmd repin --json` interface.

### IMP-repin-ApplyReport-fe50f25: ApplyReport
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: type
trace: []
```

ApplyReport describes the outcome of a repin apply step. Applied
is the number of ref replacements actually written to disk (a
delta whose SourceRef wasn't found in any reqmd block is a no-op
and not counted). Skipped holds the predated deltas, which repin
never auto-fixes. Files is the number of distinct spec files
actually rewritten.

### IMP-repin-Build-e76efe3: Build
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 58
x-reqmd.symbol-kind: function
trace: []
```

Build parses the spec tree at root and returns the list of
version-pin changes that repin would propose (or apply, when
followed by Apply). When promoteUnpinned is true, refs without
a pin against a versioned upstream are also proposed for pinning
— converting "no claim" into "claimed at current version".

The returned Result is sorted (delegated to graph.RepinDeltas).

### IMP-repin-Apply-5d0fd9b: Apply
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 102
x-reqmd.symbol-kind: function
trace:
  - SW-REPIN-001
```

reqmd:trace SW-REPIN-001
Apply rewrites each file containing at least one outdated or
unpinned delta. Predated deltas are reported back in the Skipped
slice and never applied — they are data-integrity errors that
require manual review.

The rewrite is scoped to ```reqmd blocks: each block's contents
are scanned for the SourceRef substring, which is replaced with
the new ref form (REF~newPin, or REF~newPin if the ref was
previously unpinned). All other bytes of the file are preserved
verbatim.

Files are read and rewritten relative to root through an os.Root, so a
delta path can never escape the requirements tree. If a write fails the
function returns the error; the apply report lists every delta that was
successfully applied.

### IMP-repin-applyToFile-396fcb0: applyToFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 174
x-reqmd.symbol-kind: function
trace: []
```

applyToFile rewrites the reqmd-block portions of src so that each
delta's SourceRef is replaced with the new ref form. A delta is
considered a no-op if its SourceRef isn't found in any reqmd block
(or every match is at a non-boundary position); in that case the
file is returned unchanged.

Returns the (possibly rewritten) bytes and the total number of
ref replacements actually made across all reqmd blocks.

Matching is whole-ref: SourceRef "UP-001" only matches when the
ref is at a list-element boundary (followed by `,`, `]`, or
whitespace) AND is not the prefix of a longer ref (not followed
by `~` for the unpinned-promote case). This prevents a naive
substring match from corrupting an already-pinned ref like
`UP-001~3` when the same ref is being promoted from unpinned.

### IMP-repin-refPair-3f3a0d2: refPair
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 204
x-reqmd.symbol-kind: type
trace: []
```

refPair is a pre-computed (oldRef, newRef) pair ready for byte
matching. oldRef and newRef are []byte (not strings) to avoid
repeated []byte() conversions inside the hot scan loop.

### IMP-repin-buildPairs-b1a95f0: buildPairs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 210
x-reqmd.symbol-kind: function
trace: []
```

buildPairs projects a slice of RepinDelta into refPairs, dropping
any delta whose old ref already equals the new ref (no-op).

### IMP-repin-applyPairsToBlock-d1dd176: applyPairsToBlock
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 235
x-reqmd.symbol-kind: function
trace: []
```

applyPairsToBlock scans block and emits a new []byte with every
matching ref replaced. For each position it finds the earliest
next match across all pairs, copies the unmatched run verbatim,
then emits the replacement. Returns the new bytes and the number
of replacements made. When no pair matches, returns block
unchanged (zero allocation) and a count of 0.

Cost note: this is N sub-scans into one output buffer, where N is
len(pairs) — not a true single pass. For the typical repin
workload (a handful of deltas, small reqmd blocks) this is fine;
an Aho-Corasick automaton would be the single-pass alternative but
is overkill here.

### IMP-repin-isRefBoundary-7978e98: isRefBoundary
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 292
x-reqmd.symbol-kind: function
trace: []
```

isRefBoundary reports whether the match at [start, end) inside buf
is a whole ref (not a substring of a longer token). Both edges are
checked:
   - The byte before the match must be a list-element starter (`[`,
     `,`, whitespace, or start-of-block) so we don't match "UP-001"
     inside "XUP-001".
   - The byte after the match must be a list-element terminator (` `,
     `\t`, `\n`, `\r`, `,`, `]`, or end-of-block) so we don't match
     "UP-001" inside "UP-001X".
   - The byte after must NOT be `~` — that would mean we matched the
     prefix of a longer ref (e.g. "UP-001" inside "UP-001~3"), which
     must never be rewritten.

### IMP-repin-newRefFrom-b432132: newRefFrom
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 324
x-reqmd.symbol-kind: function
trace: []
```

newRefFrom computes the new source-form ref string for a delta.
The transform is:
   - "doc-id/ID~1" + newPin=3 → "doc-id/ID~3"
   - "ID~1"       + newPin=3 → "ID~3"
   - "doc-id/ID"  + newPin=3 → "doc-id/ID~3"  (unpinned promote)
   - "ID"         + newPin=3 → "ID~3"          (unpinned promote)

### IMP-repin-attrRange-a27fc5f: attrRange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 331
x-reqmd.symbol-kind: type
trace: []
```

attrRange is an inclusive [start, end) byte range covering the
inner YAML of a reqmd block (NOT including the ``` fences).

### IMP-repin-locateAttrBlocks-482068e: locateAttrBlocks
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 342
x-reqmd.symbol-kind: function
trace: []
```

locateAttrBlocks scans src for ```reqmd\n ... \n``` fences and
returns the inner-YAML byte ranges. Blank lines between the
opening fence and the first YAML line are skipped over by the
parser; the same applies here: we treat the byte immediately
after the opening fence's newline as the range start.

The match is anchored to the literal string "```reqmd" followed
by an optional space/info segment. This matches the parser's
attrFenceInfo convention exactly.

### IMP-repin-bytesSliceReplace-0712eaa: bytesSliceReplace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 382
x-reqmd.symbol-kind: function
trace: []
```

bytesSliceReplace returns a new []byte with src[start:end]
replaced by replacement. Single allocation via nested appends.

### IMP-repin-stripPinForDisplay-6fee437: stripPinForDisplay
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 388
x-reqmd.symbol-kind: function
trace: []
```

stripPinForDisplay removes a trailing ~N from a ref so the dry-run
diff reads naturally. Delegates to model.StripPin.

### IMP-repin-FormatText-ef8de19: FormatText
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 399
x-reqmd.symbol-kind: function
trace: []
```

FormatText renders the Result as the human-readable change list
shown by `reqmd repin` in dry-run mode. The format is:

	spec/path/to/file.md  REQ-ID  REF → REF~newPin
	...
	N outdated + M unpinned + P predated refs across K files.

### IMP-repin-FormatTextApplied-c2e84fe: FormatTextApplied
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin.go
x-reqmd.source-line: 420
x-reqmd.symbol-kind: function
trace: []
```

FormatTextApplied renders an ApplyReport as a one-line summary.
Used after a successful --yes run.

## internal/repin/repin\_test.go
### IMP-repin-writeFile-e36944a: writeFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 21
x-reqmd.symbol-kind: function
trace: []
```

Helpers
writeFile creates a temp dir, writes content, and returns the dir path
(caller is responsible for cleanup via t.TempDir()).

### IMP-repin-minimalSchema-67f296b: minimalSchema
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 35
x-reqmd.symbol-kind: const
trace: []
```

minimalSchema is a permissive JSON Schema for reqmd blocks. Tests
only need the schema to exist (parser requires it); they don't
validate the attrs strictly.

### IMP-repin-makeTree-1cd1113: makeTree
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 43
x-reqmd.symbol-kind: function
trace: []
```

makeTree creates a two-doc spec tree with one upstream (versioned
at `upVersion`) and one downstream that traces to it with the
supplied pin (or no pin when pin is ""). The downstream's trace
line is rendered exactly as supplied so we can assert byte-level
edits in Apply.

### IMP-repin-TestBuild\_Outdated-1e31778: TestBuild\_Outdated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 60
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Build-e76efe3
```

Build

### IMP-repin-TestBuild\_PromoteUnpinned-9dcb605: TestBuild\_PromoteUnpinned
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 75
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Build-e76efe3
```

### IMP-repin-TestBuild\_NoPromoteUnpinned-0171e10: TestBuild\_NoPromoteUnpinned
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 89
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Build-e76efe3
```

### IMP-repin-TestBuild\_CurrentIsNoOp-e4931f1: TestBuild\_CurrentIsNoOp
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 100
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Build-e76efe3
```

### IMP-repin-TestApply\_RewritesAttrBlockOnly-7f89609: TestApply\_RewritesAttrBlockOnly
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 115
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

Apply: file rewrite semantics

### IMP-repin-TestApply\_PromoteAddsPin-026610f: TestApply\_PromoteAddsPin
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 152
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

### IMP-repin-TestApply\_PreservesProseWithTilde-cc05272: TestApply\_PreservesProseWithTilde
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 170
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

### IMP-repin-TestApply\_Idempotent-be9805f: TestApply\_Idempotent
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 212
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

### IMP-repin-TestApply\_UnpinnedPromoteDoesNotCorruptPinnedSiblings-c1fd101: TestApply\_UnpinnedPromoteDoesNotCorruptPinnedSiblings
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 237
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

TestApply_UnpinnedPromoteDoesNotCorruptPinnedSiblings guards
against a real bug discovered in the smoke test: a naive substring
match for SourceRef="UP-001" would rewrite the prefix of an
already-pinned ref "UP-001~3" to "UP-001~3~3". The apply path
must use whole-ref matching so an unpinned ref never corrupts a
pinned sibling.

### IMP-repin-TestApply\_SkipsPredated-6633c33: TestApply\_SkipsPredated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 295
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

### IMP-repin-TestApply\_MultipleDeltasSameFile-33d073f: TestApply\_MultipleDeltasSameFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 321
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

### IMP-repin-TestNewRefFrom-124c532: TestNewRefFrom
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 371
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-newRefFrom-b432132
```

newRefFrom

### IMP-repin-TestLocateAttrBlocks-c978a8e: TestLocateAttrBlocks
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 396
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-locateAttrBlocks-482068e
```

locateAttrBlocks

### IMP-repin-TestLocateAttrBlocks\_NoBlocks-ce0a15e: TestLocateAttrBlocks\_NoBlocks
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 425
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-locateAttrBlocks-482068e
```

### IMP-repin-TestLocateAttrBlocks\_IgnoresFencedButNotAttr-4a508d6: TestLocateAttrBlocks\_IgnoresFencedButNotAttr
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 432
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-locateAttrBlocks-482068e
```

### IMP-repin-TestFormatText\_Empty-aedec57: TestFormatText\_Empty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 445
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-FormatText-ef8de19
```

FormatText

### IMP-repin-TestFormatText\_Outdated-6b59b62: TestFormatText\_Outdated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 452
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-FormatText-ef8de19
```

### IMP-repin-TestResultJSON\_Roundtrip-82619ce: TestResultJSON\_Roundtrip
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 471
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Result-e706624
```

JSON output

### IMP-repin-TestBuild\_SortedByFile-d604ba6: TestBuild\_SortedByFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 507
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Build-e76efe3
```

Sort order — repin reuses graph's sort; this is a regression test
to ensure Build preserves it.

### IMP-repin-TestApply\_PreservesFileMode-ff4fcfc: TestApply\_PreservesFileMode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/repin/repin_test.go
x-reqmd.source-line: 534
x-reqmd.symbol-kind: function
trace:
  - IMP-repin-Apply-5d0fd9b
```

TestApply_PreservesFileMode verifies that rewriting a spec file keeps its
existing permissions instead of forcing a default mode.

