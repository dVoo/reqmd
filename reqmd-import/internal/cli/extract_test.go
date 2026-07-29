package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	_ "reqmd-import/internal/lang/go" // register Go plugin via init()
)

// newTestCmd returns a minimal cobra command suitable for passing to
// runExtract as the output sink.
func newTestCmd() *cobra.Command {
	return &cobra.Command{}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunExtract_HappyPath(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// Foo does foo.
func Foo() {}

// Bar is a method.
type Bar struct{}

// Baz is a method on Bar.
func (Bar) Baz() {}
`)

	f := &extractFlags{idPrefix: "IMP-"}
	cmd := newTestCmd()
	if err := runExtract(cmd, []string{src, out}, f); err != nil {
		t.Fatalf("runExtract: %v", err)
	}

	// Top-level schema written.
	if _, err := os.Stat(filepath.Join(out, "schema.yaml")); err != nil {
		t.Errorf("top-level schema.yaml missing: %v", err)
	}
	// Per-package schema + package.md written.
	pkgDir := filepath.Join(out, "alpha")
	if _, err := os.Stat(filepath.Join(pkgDir, "schema.yaml")); err != nil {
		t.Errorf("per-package schema.yaml missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "package.md")); err != nil {
		t.Errorf("package.md missing: %v", err)
	}

	// package.md should contain a Baz method under Bar with ParentID.
	md, err := os.ReadFile(filepath.Join(pkgDir, "package.md"))
	if err != nil {
		t.Fatal(err)
	}
	mdStr := string(md)
	if !strings.Contains(mdStr, "alpha-Foo") {
		t.Errorf("package.md missing Foo requirement; got:\n%s", mdStr)
	}
	if !strings.Contains(mdStr, "alpha-Bar") {
		t.Errorf("package.md missing Bar requirement; got:\n%s", mdStr)
	}
	if !strings.Contains(mdStr, "alpha-Baz") {
		t.Errorf("package.md missing Baz method requirement; got:\n%s", mdStr)
	}
	// Baz should be a child (###) of Bar, not a top-level (##).
	if !strings.Contains(mdStr, "### IMP-alpha-Baz") {
		t.Errorf("Baz should render as ### (child of Bar); got:\n%s", mdStr)
	}
}

func TestRunExtract_RejectsMissingSource(t *testing.T) {
	f := &extractFlags{idPrefix: "IMP-"}
	cmd := newTestCmd()
	err := runExtract(cmd, []string{"/nonexistent/path/xyz", t.TempDir()}, f)
	if err == nil {
		t.Fatal("expected error for missing source directory")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should mention 'does not exist'; got: %v", err)
	}
}

func TestRunExtract_RejectsFileAsSource(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "file.go", "package x\n")
	f := &extractFlags{idPrefix: "IMP-"}
	cmd := newTestCmd()
	err := runExtract(cmd, []string{filepath.Join(src, "file.go"), t.TempDir()}, f)
	if err == nil {
		t.Fatal("expected error for file (not directory) source")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error should mention 'not a directory'; got: %v", err)
	}
}

func TestRunExtract_RejectsUnknownLang(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", "package a\nfunc X() {}\n")
	out := t.TempDir()
	f := &extractFlags{idPrefix: "IMP-", lang: "cobol"}
	cmd := newTestCmd()
	err := runExtract(cmd, []string{src, out}, f)
	if err == nil {
		t.Fatal("expected error for unknown --lang")
	}
	if !strings.Contains(err.Error(), "unknown --lang") {
		t.Errorf("error should mention 'unknown --lang'; got: %v", err)
	}
}

func TestRunExtract_SkipsNoiseDirs(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "keep.go", "package keep\nfunc Keep() {}\n")
	if err := os.MkdirAll(filepath.Join(src, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(src, ".git"), "ignore.go", "package ignore\nfunc Ignore() {}\n")
	if err := os.MkdirAll(filepath.Join(src, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(src, "node_modules"), "noise.go", "package noise\nfunc Noise() {}\n")

	out := t.TempDir()
	f := &extractFlags{idPrefix: "IMP-"}
	cmd := newTestCmd()
	if err := runExtract(cmd, []string{src, out}, f); err != nil {
		t.Fatalf("runExtract: %v", err)
	}
	// The noise packages should not appear.
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() == "ignore" || e.Name() == "noise" {
			t.Errorf("noise directory %q should have been skipped", e.Name())
		}
	}
}

func TestRunExtract_LangFilter(t *testing.T) {
	src := t.TempDir()
	// .txt has no plugin; should be skipped and counted.
	writeFile(t, src, "readme.txt", "not a source file\n")
	writeFile(t, src, "a.go", "package a\nfunc X() {}\n")

	out := t.TempDir()
	f := &extractFlags{idPrefix: "IMP-"}
	cmd := newTestCmd()
	if err := runExtract(cmd, []string{src, out}, f); err != nil {
		t.Fatalf("runExtract: %v", err)
	}
	// a package should exist, readme should not.
	if _, err := os.Stat(filepath.Join(out, "a")); err != nil {
		t.Errorf("package 'a' should exist; err: %v", err)
	}
	// The summary should report the .txt skip. Capture stdout via a real cobra cmd.
}

func TestRunExtract_Deterministic(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", "package a\nfunc A() {}\nfunc B() {}\n")
	writeFile(t, src, "b.go", "package a\nfunc C() {}\n")

	out1 := t.TempDir()
	out2 := t.TempDir()
	f1 := &extractFlags{idPrefix: "IMP-"}
	f2 := &extractFlags{idPrefix: "IMP-"}
	cmd1, cmd2 := newTestCmd(), newTestCmd()
	if err := runExtract(cmd1, []string{src, out1}, f1); err != nil {
		t.Fatal(err)
	}
	if err := runExtract(cmd2, []string{src, out2}, f2); err != nil {
		t.Fatal(err)
	}
	// The per-package package.md should be byte-identical.
	md1, err := os.ReadFile(filepath.Join(out1, "a", "package.md"))
	if err != nil {
		t.Fatal(err)
	}
	md2, err := os.ReadFile(filepath.Join(out2, "a", "package.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(md1, md2) {
		t.Errorf("package.md not deterministic across runs:\nrun1:\n%s\nrun2:\n%s", md1, md2)
	}
}

func TestExtSet(t *testing.T) {
	got := extSet([]string{".go", ".py"})
	if !got[".go"] || !got[".py"] || len(got) != 2 {
		t.Errorf("extSet = %v", got)
	}
}
