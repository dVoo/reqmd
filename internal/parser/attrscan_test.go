package parser

import (
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
	}
	for i, src := range cases {
		got, err := decodeAttrYAML(src)
		if err != nil {
			t.Fatalf("case %d: scanner error: %v", i, err)
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
