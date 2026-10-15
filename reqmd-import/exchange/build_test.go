package exchange

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func decode(t *testing.T, s string) *Bundle {
	t.Helper()
	b, err := Decode(strings.NewReader(s))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	return b
}

func files(t *testing.T, b *Bundle) map[string]string {
	t.Helper()
	out, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	m := map[string]string{}
	for _, f := range out {
		m[f.Path] = string(f.Data)
	}
	return m
}

const happyBundle = `{
  "format": "reqmd-import/v1",
  "documents": [
    {
      "id-prefix": "IMP-",
      "level": "imported-requirements",
      "schema-id": "https://reqmd.dev/schemas/imported-v1.json",
      "schema-title": "Imported",
      "upstream": {"sources": [".."]}
    },
    {
      "dir": "brakes",
      "file": "package.md",
      "title": "Brakes Module",
      "notice": "Generated. Do not edit.",
      "id-prefix": "DOORS-",
      "level": "stakeholder-requirements",
      "properties": {"priority": {"type": "string", "enum": ["High", "Low"]}},
      "required": ["priority"],
      "requirements": [
        {
          "native-id": "17",
          "title": "Brake request",
          "body": "The pedal shall request braking.",
          "attrs": {"priority": "High", "x-reqmd.external-id": 17}
        },
        {
          "native-id": "42",
          "title": "Brake pressure",
          "parent": "17",
          "native-trace": ["17"],
          "trace": ["SYS-001"],
          "attrs": {"priority": "Low"}
        }
      ]
    }
  ]
}`

func TestBuildHappyPath(t *testing.T) {
	got := files(t, decode(t, happyBundle))

	// Root document: schema only.
	if _, ok := got["brakes/requirements.md"]; ok {
		t.Error("File was set, expected package.md")
	}
	schema, ok := got["schema.yaml"]
	if !ok {
		t.Fatalf("missing root schema.yaml; got %v", keys(got))
	}
	if !strings.Contains(schema, "id-prefix: IMP-") {
		t.Errorf("root schema missing id-prefix; got:\n%s", schema)
	}
	if !strings.Contains(schema, "sources:") {
		t.Errorf("root schema missing upstream sources; got:\n%s", schema)
	}

	pkgSchema, ok := got["brakes/schema.yaml"]
	if !ok {
		t.Fatalf("missing brakes/schema.yaml; got %v", keys(got))
	}
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(strings.TrimPrefix(pkgSchema, "<!-- reqmd: generated -->\n")), &parsed); err != nil {
		t.Fatalf("brakes schema is not valid YAML: %v\n%s", err, pkgSchema)
	}
	x, _ := parsed["x-reqmd"].(map[string]any)
	if x["id-prefix"] != "DOORS-" || x["level"] != "stakeholder-requirements" {
		t.Errorf("x-reqmd = %v", x)
	}
	if _, ok := parsed["required"]; !ok {
		t.Error("required list was dropped from the schema")
	}
	if _, ok := parsed["properties"]; !ok {
		t.Error("properties were dropped from the schema")
	}

	md, ok := got["brakes/package.md"]
	if !ok {
		t.Fatalf("missing brakes/package.md; got %v", keys(got))
	}
	for _, want := range []string{
		"# Brakes Module",
		"<!-- reqmd: generated -->",
		"Generated. Do not edit.",
		"## DOORS-17: Brake request",
		"### DOORS-42: Brake pressure",
		"priority: High",
		"x-reqmd.external-id: 17",
		"trace:\n  - SYS-001\n  - DOORS-17\n",
		"The pedal shall request braking.",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("package.md missing %q; got:\n%s", want, md)
		}
	}
	if !strings.Contains(md, "status: approved") {
		t.Errorf("status should default to approved; got:\n%s", md)
	}
	// Explicit trace entries come before resolved native-trace ones.
	if strings.Index(md, "- SYS-001") > strings.Index(md, "- DOORS-17") {
		t.Error("literal trace should render before resolved native-trace")
	}
}

func TestBuildFormatMismatch(t *testing.T) {
	b := decode(t, `{"format":"reqmd-import/v2","documents":[{"id-prefix":"X-"}]}`)
	if _, err := b.Build(); err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("want unsupported-format error, got %v", err)
	}
}

func TestBuildNoDocuments(t *testing.T) {
	b := decode(t, `{"format":"reqmd-import/v1","documents":[]}`)
	if _, err := b.Build(); err == nil || !strings.Contains(err.Error(), "no documents") {
		t.Fatalf("want no-documents error, got %v", err)
	}
}

func TestBuildRequiresNativeIDAndPrefix(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string
	}{
		{
			"missing native-id",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"title":"a"}]}]}`,
			"no native-id",
		},
		{
			"missing id-prefix",
			`{"format":"reqmd-import/v1","documents":[{"requirements":[{"native-id":"1"}]}]}`,
			"id-prefix is required",
		},
		{
			"duplicate native-id",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1"},{"native-id":"1"}]}]}`,
			"not unique",
		},
		{
			"unknown parent",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","parent":"9"}]}]}`,
			"parent",
		},
		{
			"unknown native-trace",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","native-trace":["9"]}]}]}`,
			"native-trace",
		},
		{
			"reserved attr",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","attrs":{"trace":["a"]}}]}]}`,
			"reserved",
		},
		{
			"escaping dir",
			`{"format":"reqmd-import/v1","documents":[{"dir":"../outside","id-prefix":"X-","requirements":[{"native-id":"1"}]}]}`,
			"escapes the output root",
		},
		{
			"duplicate output path",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1"}]},{"id-prefix":"Y-","requirements":[{"native-id":"2"}]}]}`,
			"claimed by both",
		},
		{
			"self-parent cycle",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","parent":"1"}]}]}`,
			"cycle",
		},
		{
			"two-node cycle",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","parent":"2"},{"native-id":"2","parent":"1"}]}]}`,
			"cycle",
		},
		{
			"file outside the document dir",
			`{"format":"reqmd-import/v1","documents":[{"dir":"d","file":"../../outside/x.md","id-prefix":"X-","requirements":[{"native-id":"1"}]}]}`,
			"plain file name",
		},
		{
			"absolute file",
			`{"format":"reqmd-import/v1","documents":[{"file":"/tmp/x.md","id-prefix":"X-","requirements":[{"native-id":"1"}]}]}`,
			"plain file name",
		},
		{
			"document title with a line break",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","title":"Real\n## X-9: Injected","requirements":[{"native-id":"1"}]}]}`,
			"line break",
		},
		{
			"requirement title with a line break",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","title":"One\n## X-9: Injected"}]}]}`,
			"line break",
		},
		{
			"id with whitespace",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"Item 42"}]}]}`,
			"must not contain whitespace",
		},
		{
			"attr key with a colon",
			`{"format":"reqmd-import/v1","documents":[{"id-prefix":"X-","requirements":[{"native-id":"1","attrs":{"weird:key":"v"}}]}]}`,
			"may only contain",
		},
		{
			"body with an attr fence under a heading",
			"{\"format\":\"reqmd-import/v1\",\"documents\":[{\"id-prefix\":\"X-\",\"requirements\":[{\"native-id\":\"1\",\"body\":\"text\\n## X-9: Injected\\n```attr\\nstatus: approved\\n```\"}]}]}",
			"attr fence",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := decode(t, c.json).Build()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want error containing %q, got %v", c.want, err)
			}
		})
	}
}

// TestBuildAllowsAttrFenceInProse pins the body rule: only an attr fence
// directly under a heading is rejected, because only that pair becomes a
// requirement. A body that merely mentions a fence — a doc comment documenting
// the attr format, say — must keep working.
func TestBuildAllowsAttrFenceInProse(t *testing.T) {
	got := files(t, decode(t, "{\"format\":\"reqmd-import/v1\",\"documents\":[{\"id-prefix\":\"X-\",\"requirements\":[{\"native-id\":\"1\",\"title\":\"T\",\"body\":\"Renders the heading plus its ```attr YAML block.\"}]}]}"))
	if !strings.Contains(got["requirements.md"], "plus its ```attr YAML block.") {
		t.Errorf("a body with no heading+fence pair should be kept; got:\n%s", got["requirements.md"])
	}
}

// TestBuildMovesChildrenAfterParents verifies that a child listed before its
// parent still renders nested under it, so the ## / ### heading order is valid.
func TestBuildMovesChildrenAfterParents(t *testing.T) {
	got := files(t, decode(t, `{"format":"reqmd-import/v1","documents":[
	  {"dir":"p","id-prefix":"P-","requirements":[
	    {"native-id":"m1","title":"M1","parent":"a"},
	    {"native-id":"a","title":"A"},
	    {"native-id":"b","title":"B"}
	  ]}
	]}`))
	md := got["p/requirements.md"]
	if strings.Index(md, "## P-a: A") > strings.Index(md, "### P-m1: M1") {
		t.Errorf("child should render after its parent; got:\n%s", md)
	}
	if strings.Index(md, "### P-m1: M1") > strings.Index(md, "## P-b: B") {
		t.Errorf("child should render directly after its parent, before the next top-level requirement; got:\n%s", md)
	}
}

// TestBuildSupportsMultiLevelHierarchy verifies arbitrary nesting depth: each
// level adds one '#' to the heading.
func TestBuildSupportsMultiLevelHierarchy(t *testing.T) {
	got := files(t, decode(t, `{"format":"reqmd-import/v1","documents":[
	  {"dir":"p","id-prefix":"P-","requirements":[
	    {"native-id":"1","title":"L0"},
	    {"native-id":"2","title":"L1","parent":"1"},
	    {"native-id":"3","title":"L2","parent":"2"}
	  ]}
	]}`))
	md := got["p/requirements.md"]
	for _, want := range []string{"## P-1: L0", "### P-2: L1", "#### P-3: L2"} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q; got:\n%s", want, md)
		}
	}
}

// TestBuildResolvesNativeTraceAcrossDocuments verifies that a native-trace ref
// to an item in another document resolves through that document's prefix.
func TestBuildResolvesNativeTraceAcrossDocuments(t *testing.T) {
	got := files(t, decode(t, `{
	  "format": "reqmd-import/v1",
	  "documents": [
	    {"dir":"up","id-prefix":"UP-","requirements":[{"native-id":"7","title":"U"}]},
	    {"dir":"down","id-prefix":"DN-","requirements":[{"native-id":"9","title":"D","native-trace":["7"]}]}
	  ]
	}`))
	if !strings.Contains(got["down/requirements.md"], "trace:\n  - UP-7\n") {
		t.Errorf("native-trace did not resolve to the other document's prefix; got:\n%s", got["down/requirements.md"])
	}
}

func TestDecodeRejectsUnknownField(t *testing.T) {
	_, err := Decode(strings.NewReader(`{"format":"reqmd-import/v1","documents":[],"typo":1}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("want unknown-field error, got %v", err)
	}
}

func TestDecodeRejectsTrailingData(t *testing.T) {
	_, err := Decode(strings.NewReader(`{"format":"reqmd-import/v1","documents":[]}{}`))
	if err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Fatalf("want trailing-data error, got %v", err)
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
