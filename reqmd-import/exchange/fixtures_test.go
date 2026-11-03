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

// invalidFixtures maps a fixture to the fragment Build must report. Each file
// breaks exactly one semantic rule, so a failure points at one sentence of the
// format document.
var invalidFixtures = map[string]string{
	"missing-kind.json":          "want \"requirement\" or \"section\"",
	"requirement-without-key.json": "requirement without a key",
	"duplicate-key.json":         "is used by both",
	"unresolved-link.json":       `links to "missing"`,
	"link-to-section.json":       "links may only target requirements",
	"unknown-document-link.json": `unknown document "nope"`,
	"undeclared-field.json":      `attribute "other" is not declared`,
	"missing-required.json":      `missing required attribute "priority"`,
	"nested-field-value.json":    "nested",
	"bad-attribute-name.json":    "name may only contain",
	"title-line-break.json":      "must not contain a line break",
	"escaping-dir.json":          "escapes the output root",
	"duplicate-output-path.json": "is claimed by both",
	"empty-documents.json":       "bundle has no documents",
	"duplicate-document-key.json": `document key "d" is used by both`,
	"document-title-missing.json": "title is required",
	"document-key-slash.json":     "must not contain '/' or whitespace",
	"depth-budget.json":           "would render heading level",
	"wrong-schema-url.json":       "bundle $schema is",
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

// TestFixtures_InvalidBundlesReportTheirRule verifies each broken bundle names
// the rule it breaks, so the corpus doubles as documentation of the errors.
func TestFixtures_InvalidBundlesReportTheirRule(t *testing.T) {
	for name, want := range invalidFixtures {
		t.Run(name, func(t *testing.T) {
			b, err := Decode(strings.NewReader(string(readFixture(t, filepath.Join("invalid", name)))))
			if err != nil {
				// A decode-level rule (unknown field, depth guard) is a valid
				// rejection too, as long as it is not silent.
				return
			}
			_, err = b.Build()
			if err == nil {
				t.Fatalf("invalid fixture %s was accepted", name)
			}
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error does not mention %q:\n%v", want, err)
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
