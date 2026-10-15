package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	_ "reqmd-code/internal/lang/c" // register C plugin via init()
	_ "reqmd-code/internal/lang/cpp"
	_ "reqmd-code/internal/lang/go" // register Go plugin via init()
	_ "reqmd-code/internal/lang/python"
	_ "reqmd-code/internal/lang/rust"

	"reqmd-import/exchange"
)

// runTestExtract runs an extraction and returns the decoded bundle plus the
// captured stdout/stderr text.
func runTestExtract(t *testing.T, args []string, f *extractFlags) (*exchange.Bundle, string, string) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	if err := runExtract(cmd, args, f); err != nil {
		t.Fatalf("runExtract: %v (stderr: %s)", err, errBuf.String())
	}
	bundle, err := exchange.Decode(bytes.NewReader(outBuf.Bytes()))
	if err != nil {
		t.Fatalf("decode emitted bundle: %v\nraw output:\n%s", err, outBuf.String())
	}
	return bundle, outBuf.String(), errBuf.String()
}

// renderTo renders the bundle through reqmd-import's own build path — the
// same code the real pipeline runs after the JSON hop — and returns the
// rendered files keyed by slash-separated path.
func renderTo(t *testing.T, bundle *exchange.Bundle) map[string]string {
	t.Helper()
	files, err := bundle.Build()
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}
	got := map[string]string{}
	for _, f := range files {
		got[filepath.ToSlash(f.Path)] = string(f.Data)
	}
	return got
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// docByDir returns the bundle document with the given Dir.
func docByDir(t *testing.T, b *exchange.Bundle, dir string) exchange.Document {
	t.Helper()
	for _, d := range b.Documents {
		if d.Dir == dir {
			return d
		}
	}
	t.Fatalf("bundle has no document with dir %q (%d documents)", dir, len(b.Documents))
	return exchange.Document{}
}

func TestRunExtract_HappyPath(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// Foo does foo.
func Foo() {}

// Bar is a method.
type Bar struct{}

// Baz is a method on Bar.
func (Bar) Baz() {}
`)

	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})

	// Root document: umbrella schema only, no requirements.
	root := docByDir(t, bundle, "")
	if len(root.Requirements) != 0 {
		t.Errorf("root document should carry no requirements, got %d", len(root.Requirements))
	}
	if root.IDPrefix != "IMP-" {
		t.Errorf("root id-prefix = %q, want IMP-", root.IDPrefix)
	}

	// Package document.
	alpha := docByDir(t, bundle, "alpha")
	if alpha.IDPrefix != "IMP-alpha-" {
		t.Errorf("alpha id-prefix = %q, want IMP-alpha-", alpha.IDPrefix)
	}
	if alpha.Title != "Package: alpha" || alpha.File != "package.md" {
		t.Errorf("alpha document meta = %q/%q", alpha.Title, alpha.File)
	}

	native := map[string]exchange.Requirement{}
	for _, r := range alpha.Requirements {
		native[r.Title] = r
	}
	if len(native) != 3 {
		t.Fatalf("want 3 requirements (Foo, Bar, Baz), got %d", len(native))
	}
	// Baz is a method on Bar, so it parents to Bar's native id.
	if want := native["Bar"].NativeID; native["Baz"].Parent != want {
		t.Errorf("Baz parent = %q, want %q", native["Baz"].Parent, want)
	}
	if native["Foo"].Parent != "" {
		t.Errorf("Foo should have no parent, got %q", native["Foo"].Parent)
	}

	// The rendered document is valid reqmd input.
	files := renderTo(t, bundle)
	md, ok := files["alpha/package.md"]
	if !ok {
		t.Fatalf("missing alpha/package.md; got %v", keys(files))
	}
	for _, want := range []string{"## IMP-alpha-Foo-", "## IMP-alpha-Bar-", "### IMP-alpha-Baz-", "x-reqmd.imported: true", "trace: []"} {
		if !strings.Contains(md, want) {
			t.Errorf("package.md missing %q; got:\n%s", want, md)
		}
	}
	if _, ok := files["schema.yaml"]; !ok {
		t.Errorf("root schema.yaml missing; got %v", keys(files))
	}
	if _, ok := files["alpha/schema.yaml"]; !ok {
		t.Errorf("per-package schema.yaml missing; got %v", keys(files))
	}
}

func TestRunExtract_RejectsMissingSource(t *testing.T) {
	cmd := &cobra.Command{}
	err := runExtract(cmd, []string{"/nonexistent/path/xyz"}, &extractFlags{idPrefix: "IMP-"})
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
	cmd := &cobra.Command{}
	err := runExtract(cmd, []string{filepath.Join(src, "file.go")}, &extractFlags{idPrefix: "IMP-"})
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
	cmd := &cobra.Command{}
	err := runExtract(cmd, []string{src}, &extractFlags{idPrefix: "IMP-", lang: "cobol"})
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

	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	for _, d := range bundle.Documents {
		if d.Dir == "ignore" || d.Dir == "noise" {
			t.Errorf("noise package %q should have been skipped", d.Dir)
		}
	}
}

func TestRunExtract_LangFilter(t *testing.T) {
	src := t.TempDir()
	// .txt has no plugin; should be skipped and counted.
	writeFile(t, src, "readme.txt", "not a source file\n")
	writeFile(t, src, "a.go", "package a\nfunc X() {}\n")

	bundle, _, stderr := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	docByDir(t, bundle, "a")
	if !strings.Contains(stderr, "skipped 1 file(s)") {
		t.Errorf("summary should report the .txt skip; got: %q", stderr)
	}
}

func TestRunExtract_Deterministic(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", "package a\nfunc A() {}\nfunc B() {}\n")
	writeFile(t, src, "b.go", "package a\nfunc C() {}\n")

	_, out1, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	_, out2, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	if out1 != out2 {
		t.Errorf("bundle not deterministic across runs:\nrun1:\n%s\nrun2:\n%s", out1, out2)
	}
}

func TestExtSet(t *testing.T) {
	got := extSet([]string{".go", ".py"})
	if !got[".go"] || !got[".py"] || len(got) != 2 {
		t.Errorf("extSet = %v", got)
	}
}

// TestRunExtract_NoSymbolsNotParseFailure verifies that a file with no
// extractable symbols (e.g. a package-doc-only .go file) is not counted
// as a parse failure in the summary.
func TestRunExtract_NoSymbolsNotParseFailure(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "doc.go", "// Package only.\npackage doconly\n")

	_, _, stderr := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	if strings.Contains(stderr, "parse failure") {
		t.Errorf("summary should not report parse failures; got: %q", stderr)
	}
}

// TestRunExtract_MultiLanguage walks a source tree containing one file per
// registered language and verifies that every language plugin self-registers
// and claims its extension.
func TestRunExtract_MultiLanguage(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "thing.c", `// C thing.
int c_thing(void) { return 0; }
`)
	writeFile(t, src, "thing.cpp", `// C++ thing.
int cpp_thing() { return 0; }
`)
	writeFile(t, src, "thing.go", `package gopkg

// Go thing.
func GoThing() {}
`)
	writeFile(t, src, "thing.py", `# Python thing.
def py_thing():
    pass
`)
	writeFile(t, src, "thing.rs", `// Rust thing.
fn rust_thing() {}
`)

	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})

	// The Go package carries the package-clause name; the C/C++/Python/Rust
	// packages fall back to the file basename ("thing"), all sharing one
	// document.
	gopkg := docByDir(t, bundle, "gopkg")
	thing := docByDir(t, bundle, "thing")

	var titles []string
	for _, d := range []exchange.Document{gopkg, thing} {
		for _, r := range d.Requirements {
			titles = append(titles, r.Title)
		}
	}
	for _, want := range []string{"GoThing", "c_thing", "cpp_thing", "py_thing", "rust_thing"} {
		if !contains(titles, want) {
			t.Errorf("bundle missing %q requirement; got %v", want, titles)
		}
	}
}

// TestRunExtract_OutputFile verifies -o writes the bundle to a file and keeps
// stdout clean.
func TestRunExtract_OutputFile(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", "package a\nfunc X() {}\n")
	out := filepath.Join(t.TempDir(), "bundle.json")

	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	if err := runExtract(cmd, []string{src}, &extractFlags{idPrefix: "IMP-", output: out}); err != nil {
		t.Fatalf("runExtract: %v", err)
	}
	if outBuf.Len() != 0 {
		t.Errorf("stdout should stay empty with -o; got %q", outBuf.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if _, err := exchange.Decode(bytes.NewReader(data)); err != nil {
		t.Errorf("bundle file is not decodable: %v", err)
	}
	if !strings.Contains(errBuf.String(), "extracted") {
		t.Errorf("summary should go to stderr; got %q", errBuf.String())
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
