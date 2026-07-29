package writer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStandardSchemaFilesIdentical guards against drift between the
// canonical schema at internal/schema/standard_schema.yaml and the
// synced copy under this package (which //go:embed requires because
// go:embed cannot cross package boundaries). If a developer updates
// the canonical schema and forgets to copy it here, the embedded
// copy would silently diverge from the source-of-truth. This test
// fails fast in that case.
func TestStandardSchemaFilesIdentical(t *testing.T) {
	// Walk up from this test file's working directory to the workspace
	// root. The standard layout is:
	//   <workspace>/reqmd-import/internal/writer/<this file>
	//   <workspace>/reqmd-import/internal/schema/standard_schema.yaml  (canonical)
	//   <workspace>/reqmd-import/internal/writer/standard_schema.yaml  (embedded copy)
	// so going up 3 levels from the package dir reaches the workspace root.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repo := filepath.Clean(filepath.Join(wd, "..", "..", ".."))

	canonicalPath := filepath.Join(repo, "reqmd-import", "internal", "schema", "standard_schema.yaml")
	embeddedPath := filepath.Join(wd, "standard_schema.yaml")

	canonical, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("read canonical schema at %s: %v", canonicalPath, err)
	}
	embedded, err := os.ReadFile(embeddedPath)
	if err != nil {
		t.Fatalf("read embedded copy at %s: %v", embeddedPath, err)
	}

	if len(canonical) != len(embedded) {
		t.Errorf("schema copies differ in size: canonical=%d embedded=%d bytes",
			len(canonical), len(embedded))
		t.Logf("if you updated internal/schema/standard_schema.yaml, " +
			"copy it to internal/writer/standard_schema.yaml")
	}
	for i := 0; i < len(canonical) && i < len(embedded); i++ {
		if canonical[i] != embedded[i] {
			t.Errorf("schema copies differ at byte %d: canonical=%q embedded=%q",
				i, canonical[i], embedded[i])
			break
		}
	}
}
