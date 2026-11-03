package exchange

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

// schemaPath is the published contract this format claims to implement. The
// fixtures below are the shared corpus a third-party producer can copy: every
// valid bundle must satisfy both the schema (shape) and Build (semantics), and
// every invalid bundle names the rule it breaks.
const schemaPath = "../../site/static/exchange/v1/schema.json"

// validFixtures must pass the schema and Build.
var validFixtures = []string{
	"doors-sections.json",
	"code-proxies.json",
	"status-mapping.json",
	"cross-document-link.json",
	"schema-only.json",
}

// invalidFixtures must be rejected. Which rule each one breaks, and with which
// message, is asserted once in build_test.go: the corpus's job is that the
// published schema and the renderer agree about what is legal, not to restate
// the wording of every error.
var invalidFixtures = []string{
	"missing-kind.json",
	"requirement-without-key.json",
	"duplicate-key.json",
	"unresolved-link.json",
	"link-to-section.json",
	"unknown-document-link.json",
	"undeclared-field.json",
	"missing-required.json",
	"nested-field-value.json",
	"bad-attribute-name.json",
	"title-line-break.json",
	"escaping-dir.json",
	"duplicate-output-path.json",
	"empty-documents.json",
	"duplicate-document-key.json",
	"document-title-missing.json",
	"document-key-slash.json",
	"depth-budget.json",
	"wrong-schema-url.json",
}

// TestFixtures_SchemaAcceptsValidBundles is the neutrality check: the published
// schema alone accepts every well-formed bundle, with no Go-specific knowledge.
func TestFixtures_SchemaAcceptsValidBundles(t *testing.T) {
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read published schema: %v", err)
	}
	var s jsonschema.Schema
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("parse published schema: %v", err)
	}
	resolved, err := s.Resolve(&jsonschema.ResolveOptions{})
	if err != nil {
		t.Fatalf("resolve published schema: %v", err)
	}

	for _, name := range validFixtures {
		t.Run(name, func(t *testing.T) {
			var doc any
			if err := json.Unmarshal(readFixture(t, filepath.Join("valid", name)), &doc); err != nil {
				t.Fatalf("fixture is not valid JSON: %v", err)
			}
			if err := resolved.Validate(doc); err != nil {
				t.Errorf("published schema rejects a valid bundle: %v", err)
			}
		})
	}
}

// TestFixtures_ValidBundlesBuild verifies every valid fixture renders.
func TestFixtures_ValidBundlesBuild(t *testing.T) {
	for _, name := range validFixtures {
		t.Run(name, func(t *testing.T) {
			b, err := Decode(strings.NewReader(string(readFixture(t, filepath.Join("valid", name)))))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if _, err := b.Build(); err != nil {
				t.Errorf("valid fixture rejected: %v", err)
			}
		})
	}
}

// TestFixtures_InvalidBundlesAreRejected verifies the corpus's negative half:
// nothing in it renders. A rejection at decode time (unknown field, nesting
// guard) counts the same as one at build time — both are loud.
func TestFixtures_InvalidBundlesAreRejected(t *testing.T) {
	for _, name := range invalidFixtures {
		t.Run(name, func(t *testing.T) {
			b, err := Decode(strings.NewReader(string(readFixture(t, filepath.Join("invalid", name)))))
			if err != nil {
				return
			}
			if _, err := b.Build(); err == nil {
				t.Fatalf("invalid fixture %s was accepted", name)
			}
		})
	}
}

// readFixture loads one fixture from the corpus.
func readFixture(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", rel))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}
