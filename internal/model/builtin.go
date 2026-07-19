package model

import (
	"strconv"
	"strings"
)

// Built-in attribute keys used in requirement attr blocks. They are
// injected into every schema by internal/schema.Compile and accessed by
// graph checks and HTML export.
const (
	AttrStatus            = "status"
	AttrTrace             = "trace"
	AttrDisposition       = "disposition"
	AttrDispositionReason = "disposition-reason"
	AttrVersion           = "version"
	AttrRequiresTraceFrom = "requires-trace-from"
)

// StripPin removes a trailing `~N` version pin from a requirement ref.
// Returns the bare ref, the pin value, and true if a numeric pin was
// present. A `~` followed by a non-numeric suffix is NOT treated as a
// pin (e.g. "UP-001~abc" returns ("UP-001~abc", 0, false)) — this
// matches the graph layer's parsing and prevents mis-stripping refs
// that contain `~` for other reasons.
//
// This is the single source of truth for `~N` stripping across the
// codebase: graph (Pass 2 pin extraction), repin (newRefFrom, stripPin
// for display), and verify (MergeLatest key normalization).
func StripPin(ref string) (bare string, pin int, pinned bool) {
	idx := strings.LastIndex(ref, "~")
	if idx < 0 {
		return ref, 0, false
	}
	n, err := strconv.Atoi(ref[idx+1:])
	if err != nil {
		return ref, 0, false
	}
	return ref[:idx], n, true
}

// Built-in status enum values. The default (when no status attr is
// declared) is StatusApproved — see StatusDefault.
const (
	StatusApproved = "approved"
	StatusDraft    = "draft"
)

// StatusDefault is the implicit status when none is declared on a
// requirement. The built-in `status` attribute is enum [draft, approved];
// `approved` is the coverage-providing value.
const StatusDefault = StatusApproved
