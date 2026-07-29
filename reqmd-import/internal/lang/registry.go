// Package lang defines the Language plugin interface and a global registry
// that language implementations self-register into via init().
//
// To add a new language to reqmd-import, drop a new directory under
// internal/lang/<lang>/ with a Go file that implements Language and calls
// lang.Register(&impl{}) in an init() block. The blank import in
// cmd/reqmd-import/main.go pulls all plugins into the binary.
package lang

import (
	"fmt"
	"sort"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"

	"reqmd-import/internal/model"
)

// Language is the contract a source-language plugin must satisfy.
//
// Implementations are expected to be safe to call concurrently: the
// registry caches a single instance per language and the extractor
// dispatches jobs to language plugins in parallel.
type Language interface {
	// Name returns the canonical, lowercase language identifier
	// (e.g. "go", "python"). Used in CLI flags, log output, and the
	// reqmd-importer config schema's x-reqmd.language field.
	Name() string

	// Extensions lists the file extensions this language plugin handles,
	// including the leading dot (e.g. []string{".go"}).
	Extensions() []string

	// Grammar returns the tree-sitter grammar for this language. The
	// returned pointer is owned by the implementation; callers must not
	// close it.
	Grammar() *sitter.Language

	// Query returns the embedded tree-sitter query (in .scm syntax) used
	// to capture symbols from the parsed tree. An empty string means
	// "no symbols" (e.g. a placeholder plugin).
	Query() string

	// BindDoc extracts and returns the cleaned doc comment immediately
	// preceding the given declaration node. Returns "" if no comment is
	// present or the comment is empty after cleaning. The source bytes
	// are the full file contents; the node is a top-level declaration
	// captured by the language's Query.
	BindDoc(node *sitter.Node, src []byte) string

	// Parse parses the given source bytes and returns a flat list of
	// top-level symbols in source order. The file path is carried into
	// each Symbol for provenance.
	Parse(file string, src []byte) ([]model.Symbol, error)
}

// Registry is the global, append-only set of registered Language plugins.
// It is safe for concurrent reads after the init() phase; concurrent
// Register calls during init() are protected by a mutex.
type registry struct {
	mu        sync.RWMutex
	languages map[string]Language
}

var globalRegistry = &registry{
	languages: make(map[string]Language),
}

// Register adds a Language to the global registry. Typically called from
// a plugin's init() function. Duplicate registrations panic — they
// indicate a programming error (two packages claiming the same language).
func Register(l Language) {
	if l == nil {
		panic("lang: cannot register nil Language")
	}
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if _, exists := globalRegistry.languages[l.Name()]; exists {
		panic(fmt.Sprintf("lang: language %q already registered", l.Name()))
	}
	globalRegistry.languages[l.Name()] = l
}

// Get fetches a registered Language by name. Returns (nil, false) if
// the language is not registered.
func Get(name string) (Language, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	l, ok := globalRegistry.languages[name]
	return l, ok
}

// All returns every registered language, sorted by name for deterministic
// output. Used by `reqmd-import list-languages` (M2+) and by the
// extractor's language auto-detection pass.
func All() []Language {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	out := make([]Language, 0, len(globalRegistry.languages))
	for _, l := range globalRegistry.languages {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// ForExtension returns the first Language that claims the given file
// extension. Returns (nil, false) if no plugin handles the extension.
// Comparison is exact-match on the extensions list (no normalization).
func ForExtension(ext string) (Language, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	for _, l := range globalRegistry.languages {
		for _, e := range l.Extensions() {
			if e == ext {
				return l, true
			}
		}
	}
	return nil, false
}
