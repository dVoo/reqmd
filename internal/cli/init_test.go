package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInit_GenericPreset checks the default preset renders a schema and an
// example document with the requested prefix.
func TestInit_GenericPreset(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-spec")

	_, err := runCmd(t, newInitCmd(), dir, "--id-prefix", "REQ-")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	schema := readFileString(t, filepath.Join(dir, "schema.yaml"))
	for _, want := range []string{`level: "requirements"`, "id-prefix: REQ-", `title: "my-spec Requirements"`} {
		if !strings.Contains(schema, want) {
			t.Errorf("schema.yaml missing %q:\n%s", want, schema)
		}
	}

	example := readFileString(t, filepath.Join(dir, "requirements.md"))
	if !strings.Contains(example, "## REQ-001") {
		t.Errorf("example does not use the id-prefix:\n%s", example)
	}
}

// TestInit_ResultsPreset checks the results preset overrides level and prefix
// so the scaffolded directory is a valid manual-results document.
func TestInit_ResultsPreset(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "results")

	_, err := runCmd(t, newInitCmd(), dir, "--preset", "results")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	schema := readFileString(t, filepath.Join(dir, "schema.yaml"))
	for _, want := range []string{"level: verify-results", "id-prefix: VR", "x-reqmd.outcome"} {
		if !strings.Contains(schema, want) {
			t.Errorf("results schema missing %q:\n%s", want, schema)
		}
	}
}

// TestInit_CustomPresetDir checks a preset directory renders both templates
// with the schema variables.
func TestInit_CustomPresetDir(t *testing.T) {
	preset := t.TempDir()
	writeFile(t, preset, "schema.yaml.tmpl", "x-reqmd:\n  level: \"{{ .Level }}\"\n  id-prefix: {{ .IDPrefix }}\n")
	writeFile(t, preset, "example.md.tmpl", "# {{ .Title }}\n\n## {{ .IDPrefix }}001\n")

	target := filepath.Join(t.TempDir(), "custom")
	_, err := runCmd(t, newInitCmd(), target, "--preset", preset, "--id-prefix", "SC-", "--level", "safety")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	schema := readFileString(t, filepath.Join(target, "schema.yaml"))
	if !strings.Contains(schema, "id-prefix: SC-") || !strings.Contains(schema, `level: "safety"`) {
		t.Errorf("custom schema not rendered with variables:\n%s", schema)
	}
	example := readFileString(t, filepath.Join(target, "requirements.md"))
	if !strings.Contains(example, "SC-001") {
		t.Errorf("custom example not rendered:\n%s", example)
	}
}

// TestInit_ExistingFilesNeedForce checks an existing document aborts the run
// unless --force is given.
func TestInit_ExistingFilesNeedForce(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "existing")
	if _, err := runCmd(t, newInitCmd(), dir); err != nil {
		t.Fatalf("first init: %v", err)
	}

	if _, err := runCmd(t, newInitCmd(), dir); err == nil {
		t.Fatal("expected an error when the schema already exists")
	}

	if _, err := runCmd(t, newInitCmd(), dir, "--force"); err != nil {
		t.Fatalf("init --force: %v", err)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}
