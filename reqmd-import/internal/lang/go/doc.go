// Package golang implements the Go language plugin for reqmd-import:
// it parses a Go source file via tree-sitter and emits a flat list of
// top-level symbols (functions, methods, types, constants, variables).
//
// The directory is named "go" to keep imports short and conventional,
// but the package identifier cannot be "go" (a Go keyword), so we
// declare it as "golang". Callers import it as:
//
//	import golang "reqmd-import/internal/lang/go"
package golang
