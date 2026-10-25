package parser

import (
	"errors"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDecodeAttrYAML_Scalars(t *testing.T) {
	src := "status: approved\npriority: Critical\nversion: 3\nactive: true\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	if m["status"] != "approved" {
		t.Errorf("status = %v (%T), want \"approved\"", m["status"], m["status"])
	}
	if m["priority"] != "Critical" {
		t.Errorf("priority = %v, want \"Critical\"", m["priority"])
	}
	if m["version"] != 3 {
		t.Errorf("version = %v (%T), want 3 (int)", m["version"], m["version"])
	}
	if m["active"] != true {
		t.Errorf("active = %v (%T), want true (bool)", m["active"], m["active"])
	}
}

func TestDecodeAttrYAML_QuotedStrings(t *testing.T) {
	src := `process-name: "Software Verification"
category: 'Engineering'
`
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	if m["process-name"] != "Software Verification" {
		t.Errorf("process-name = %v, want \"Software Verification\"", m["process-name"])
	}
	if m["category"] != "Engineering" {
		t.Errorf("category = %v, want \"Engineering\"", m["category"])
	}
}

func TestDecodeAttrYAML_BlockList(t *testing.T) {
	src := "trace:\n  - STK-001\n  - STK-002\n  - STK-003\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	trace, ok := m["trace"].([]any)
	if !ok {
		t.Fatalf("trace is %T, want []any", m["trace"])
	}
	if len(trace) != 3 {
		t.Fatalf("trace len = %d, want 3", len(trace))
	}
	want := []string{"STK-001", "STK-002", "STK-003"}
	for i, w := range want {
		if trace[i] != w {
			t.Errorf("trace[%d] = %v, want %q", i, trace[i], w)
		}
	}
}

func TestDecodeAttrYAML_FlowSeq(t *testing.T) {
	src := "trace: [SYS-001, SAFE-003]\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	trace, ok := m["trace"].([]any)
	if !ok {
		t.Fatalf("trace is %T, want []any", m["trace"])
	}
	if len(trace) != 2 {
		t.Fatalf("trace len = %d, want 2", len(trace))
	}
	if trace[0] != "SYS-001" {
		t.Errorf("trace[0] = %v, want SYS-001", trace[0])
	}
	if trace[1] != "SAFE-003" {
		t.Errorf("trace[1] = %v, want SAFE-003", trace[1])
	}
}

func TestDecodeAttrYAML_EmptyValue(t *testing.T) {
	src := "status:\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	if m["status"] != nil {
		t.Errorf("status = %v, want nil", m["status"])
	}
}

func TestDecodeAttrYAML_Comments(t *testing.T) {
	src := "status: approved # this is a comment\n# full line comment\npriority: High\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	if m["status"] != "approved" {
		t.Errorf("status = %v, want \"approved\"", m["status"])
	}
	if m["priority"] != "High" {
		t.Errorf("priority = %v, want \"High\"", m["priority"])
	}
	if _, ok := m["# full line comment"]; ok {
		t.Error("comment line should be ignored")
	}
}

func TestDecodeAttrYAML_BlankLines(t *testing.T) {
	src := "status: approved\n\n\npriority: High\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(m))
	}
}

func TestDecodeAttrYAML_EmptyKey(t *testing.T) {
	_, err := decodeAttrYAML(": value\n")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestDecodeAttrYAML_BadYAML(t *testing.T) {
	_, err := decodeAttrYAML(": : invalid yaml\n")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestDecodeAttrYAML_MixedTypes(t *testing.T) {
	src := "status: approved\ntrace:\n  - STK-001\n  - STK-002\nversion: 5\nname: \"Test\"\n"
	m, err := decodeAttrYAML(src)
	if err != nil {
		t.Fatal(err)
	}
	if m["status"] != "approved" {
		t.Errorf("status = %v", m["status"])
	}
	if m["version"] != 5 {
		t.Errorf("version = %v (%T)", m["version"], m["version"])
	}
	if m["name"] != "Test" {
		t.Errorf("name = %v", m["name"])
	}
	trace := m["trace"].([]any)
	if len(trace) != 2 {
		t.Fatalf("trace len = %d", len(trace))
	}
}

func TestDecodeAttrYAML_ParityWithYAMLv3(t *testing.T) {
	cases := []string{
		"status: approved\npriority: Critical\n",
		"trace:\n  - STK-001\n  - STK-002\n",
		"trace: [A, B, C]\n",
		"version: 3\nactive: true\n",
		`name: "Software Verification"`,
		"empty: null\n",
		// Scalars yaml.v3 types as non-string: the scanner must not turn
		// them into strings (it reports errUnsupported and decodeAttrs
		// re-parses with yaml.v3).
		"weight: 1.5\n",
		"weight: -1.5\n",
		"date: 2024-01-05\n",
		"stamp: 2024-01-05T10:00:00Z\n",
		"big: 9223372036854775808\n",
		"hex: 0x1f\n",
		"inf: .inf\n",
		// Not retyped by yaml.v3, must stay strings (fast path).
		"label: v1.2\nhexish: 0xZZ\n",
	}
	for i, src := range cases {
		got, err := decodeAttrs(src)
		if err != nil {
			t.Fatalf("case %d: decode error: %v", i, err)
		}
		var want map[string]any
		if err := yaml.Unmarshal([]byte(src), &want); err != nil {
			t.Fatalf("case %d: yaml.v3 error: %v", i, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("case %d:\n  got  = %#v\n  want = %#v", i, got, want)
		}
	}
}

// TestDecodeAttrs_DuplicateKey rejects a repeated key instead of silently
// keeping the last value (yaml.v3 rejects it too, with a line number).
func TestDecodeAttrs_DuplicateKey(t *testing.T) {
	if _, err := decodeAttrs("status: draft\nstatus: approved\n"); err == nil {
		t.Fatal("expected error for duplicate key")
	}
}

// TestDecodeAttrYAML_UnsupportedScalarIsReported guards the fast path: a
// scalar it cannot type must be reported as errUnsupported so decodeAttrs
// falls back, never silently stringified.
func TestDecodeAttrYAML_UnsupportedScalarIsReported(t *testing.T) {
	if _, err := decodeAttrYAML("weight: 1.5\n"); !errors.Is(err, errUnsupported) {
		t.Fatalf("err = %v, want errUnsupported", err)
	}
}

// TestParseScalar covers the plain-scalar typing the scanner must reproduce
// exactly; anything it cannot type is rejected so the block re-parses with
// yaml.v3 instead of silently becoming a string.
func TestParseScalar(t *testing.T) {
	cases := []struct {
		in   string
		want any
		err  bool
	}{
		{in: "approved", want: "approved"},
		{in: "_underscore", want: "_underscore"},
		{in: "v1.2", want: "v1.2"},
		{in: "STK-001", want: "STK-001"},
		{in: "3", want: 3},
		{in: "-3", want: -3},
		{in: "true", want: true},
		{in: "False", want: false},
		{in: "null", want: nil},
		{in: "~", want: nil},
		{in: "1.5", err: true},
		{in: "2024-01-05", err: true},
		{in: "0x1f", err: true},
		{in: ".inf", err: true},
		{in: "", err: true},
	}
	for _, c := range cases {
		got, err := parseScalar(c.in)
		if c.err {
			if !errors.Is(err, errUnsupported) {
				t.Errorf("parseScalar(%q) err = %v, want errUnsupported", c.in, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseScalar(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseScalar(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestIsPlainStringStart(t *testing.T) {
	for _, c := range []struct {
		in   byte
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'_', true},
		{'0', false},
		{'-', false},
		{'+', false},
		{'.', false},
		{'~', false},
	} {
		if got := isPlainStringStart(c.in); got != c.want {
			t.Errorf("isPlainStringStart(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIsIntegerString(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{"0", true},
		{"42", true},
		{"-7", true},
		{"+7", true},
		{"", false},
		{"-", false},
		{"1.5", false},
		{"12a", false},
	} {
		if got := isIntegerString(c.in); got != c.want {
			t.Errorf("isIntegerString(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestUnquoteYAML covers both quote styles: doubled single quotes are a
// literal quote, double quotes carry YAML escapes, and an unterminated quote
// is an error (which send the block to yaml.v3).
func TestUnquoteYAML(t *testing.T) {
	cases := []struct {
		in, want string
		err      bool
	}{
		{in: `"plain"`, want: "plain"},
		{in: `'it''s'`, want: "it's"},
		{in: `"tab\there"`, want: "tab\there"},
		{in: `"unknown\q"`, want: "unknownq"},
		{in: `"1.5"`, want: "1.5"},
		{in: `unquoted`, want: "unquoted"},
		{in: `"unterminated`, err: true},
	}
	for _, c := range cases {
		got, err := unquoteYAML(c.in)
		if c.err {
			if err == nil {
				t.Errorf("unquoteYAML(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("unquoteYAML(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("unquoteYAML(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestUnescapeYAMLDouble(t *testing.T) {
	if got := unescapeYAMLDouble(`no escapes`); got != "no escapes" {
		t.Errorf("got %q", got)
	}
	if got := unescapeYAMLDouble(`a\\b\"c\r\n`); got != "a\\b\"c\r\n" {
		t.Errorf("got %q", got)
	}
}

func TestLeadingSpaces(t *testing.T) {
	for _, c := range []struct {
		in   string
		want int
	}{
		{"", 0}, {"- item", 0}, {"  - item", 2}, {"    - item", 4}, {"\t- item", 0},
	} {
		if got := leadingSpaces(c.in); got != c.want {
			t.Errorf("leadingSpaces(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// TestParseFlowSeq covers flow sequences, including quoted items with commas
// (which a naive string split would break) and invalid input.
func TestParseFlowSeq(t *testing.T) {
	cases := []struct {
		in   string
		want []any
		err  bool
	}{
		{in: "[]", want: []any{}},
		{in: "[a, b, c]", want: []any{"a", "b", "c"}},
		{in: `["x, y", z]`, want: []any{"x, y", "z"}},
		{in: "[1, true]", want: []any{1, true}},
		{in: "not a sequence", err: true},
		{in: "[1.5]", err: true},
	}
	for _, c := range cases {
		got, err := parseFlowSeq(c.in)
		if c.err {
			if err == nil {
				t.Errorf("parseFlowSeq(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseFlowSeq(%q) unexpected error: %v", c.in, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseFlowSeq(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

// TestSplitFlowItems covers the quote- and bracket-aware comma split.
func TestSplitFlowItems(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{in: "a, b", want: []string{"a", " b"}},
		{in: `"a, b", c`, want: []string{`"a, b"`, " c"}},
		{in: "[a, b], c", want: []string{"[a, b]", " c"}},
		{in: "single", want: []string{"single"}},
	}
	for _, c := range cases {
		if got := splitFlowItems(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitFlowItems(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}
