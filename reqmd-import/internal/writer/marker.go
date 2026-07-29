package writer

import "bytes"

// GeneratedMarker is the HTML-comment sentinel embedded in the header of
// every auto-generated .md file. It is used by the writer's idempotent
// write-if-changed path to distinguish generated content from hand-authored
// content living in the same target tree.
const GeneratedMarker = "<!-- reqmd-import: generated -->"

// IsGenerated reports whether content carries the generated marker.
// Hand-authored files that happen to live next to generated output are
// left untouched on rewrite.
func IsGenerated(content []byte) bool {
	return bytes.Contains(content, []byte(GeneratedMarker))
}
