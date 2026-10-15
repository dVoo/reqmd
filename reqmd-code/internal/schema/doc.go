// Package schema holds the canonical standard_schema.yaml — the
// attribute-block schema template that emitted .md files must conform to.
// The file is consumed by internal/writer (via a local //go:embed copy,
// since go:embed cannot cross package boundaries) and by the schema-drift
// test in internal/writer, which asserts the writer's embedded copy matches
// this canonical source-of-truth.
package schema
