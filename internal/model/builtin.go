package model

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
