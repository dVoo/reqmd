package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/reqmd-import/exchange"

	"github.com/spf13/cobra"

	_ "github.com/dVoo/reqmd/reqmd-code/internal/lang/c" // register C plugin via init()
	_ "github.com/dVoo/reqmd/reqmd-code/internal/lang/cpp"
	_ "github.com/dVoo/reqmd/reqmd-code/internal/lang/go" // register Go plugin via init()
	_ "github.com/dVoo/reqmd/reqmd-code/internal/lang/python"
	_ "github.com/dVoo/reqmd/reqmd-code/internal/lang/rust"
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
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// docByDir returns the bundle document that renders into the given directory.
func docByDir(t *testing.T, b *exchange.Bundle, dir string) exchange.Document {
	t.Helper()
	for _, d := range b.Documents {
		if d.Dir() == dir {
			return d
		}
	}
	t.Fatalf("bundle has no document with dir %q (%d documents)", dir, len(b.Documents))
	return exchange.Document{}
}

// allItems flattens a document's item tree in render order.
func allItems(items []exchange.Item) []exchange.Item {
	out := make([]exchange.Item, 0, len(items))
	for _, it := range items {
		out = append(out, it)
		out = append(out, allItems(it.Children)...)
	}
	return out
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
	root := docByDir(t, bundle, ".")
	if len(root.Items) != 0 {
		t.Errorf("root document should carry no requirements, got %d", len(root.Items))
	}
	if root.XReqmd.IDPrefix != "IMP-" {
		t.Errorf("root id-prefix = %q, want IMP-", root.XReqmd.IDPrefix)
	}

	// Package document.
	alpha := docByDir(t, bundle, "alpha")
	if alpha.XReqmd.IDPrefix != "IMP-alpha-" {
		t.Errorf("alpha id-prefix = %q, want IMP-alpha-", alpha.XReqmd.IDPrefix)
	}
	if alpha.Title != "Package: alpha" || alpha.XReqmd.File != "package.md" {
		t.Errorf("alpha document meta = %q/%q", alpha.Title, alpha.XReqmd.File)
	}

	// The package document groups symbols into one section per source file.
	if len(alpha.Items) != 1 || alpha.Items[0].Kind != exchange.KindSection || alpha.Items[0].Title != "a.go" {
		t.Fatalf("alpha items = %+v, want a single section for a.go", alpha.Items)
	}
	native := map[string]exchange.Item{}
	for _, r := range allItems(alpha.Items) {
		if r.Kind == exchange.KindRequirement {
			native[r.Title] = r
		}
	}
	if len(native) != 3 {
		t.Fatalf("want 3 requirements (Foo, Bar, Baz), got %d", len(native))
	}
	// Baz is a method on Bar, so it nests as Bar's child item. The renderer
	// turns that into heading depth, so there is no parent pointer to check.
	// Children carry the doc-comment section first, then nested symbols.
	bar := native["Bar"]
	var barMembers []string
	for _, c := range bar.Children {
		if c.Kind == exchange.KindRequirement {
			barMembers = append(barMembers, c.Title)
		}
	}
	if len(barMembers) != 1 || barMembers[0] != "Baz" {
		t.Errorf("Bar requirement children = %v, want [Baz]", barMembers)
	}
	for _, c := range native["Foo"].Children {
		if c.Kind == exchange.KindRequirement {
			t.Errorf("Foo should have no nested requirements, got %q", c.Title)
		}
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
		if d.Dir() == "ignore" || d.Dir() == "noise" {
			t.Errorf("noise package %q should have been skipped", d.Dir())
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
		for _, r := range allItems(d.Items) {
			titles = append(titles, r.Title)
		}
	}
	for _, want := range []string{"GoThing", "c_thing", "cpp_thing", "py_thing", "rust_thing"} {
		if !slices.Contains(titles, want) {
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

// TestDocSections covers moving doc comments out of requirement bodies: the
// intro becomes an untitled section, every heading becomes a section, and a
// heading inside a fenced block is left alone as prose.
func TestDocSections(t *testing.T) {
	doc := "Foo does foo.\n\nUsage:\n\n## Options\n\n- fast\n\n```\n## not a heading\n```\n\n## Notes\n\nCareful."
	got := docSections(doc)
	if len(got) != 3 {
		t.Fatalf("want intro + 2 sections, got %d: %+v", len(got), got)
	}
	if got[0].Title != "" || !strings.Contains(got[0].Body, "Foo does foo.") {
		t.Errorf("intro section = %+v, want untitled prose", got[0])
	}
	if got[1].Title != "Options" || !strings.Contains(got[1].Body, "- fast") {
		t.Errorf("first section = %+v, want Options with its prose", got[1])
	}
	if got[2].Title != "Notes" {
		t.Errorf("second section title = %q, want Notes", got[2].Title)
	}
	if !strings.Contains(got[1].Body, "## not a heading") {
		t.Errorf("a fenced heading must stay in the prose, got %q", got[1].Body)
	}
	if strings.Contains(got[2].Body, "not a heading") {
		t.Errorf("the fenced block must not leak into the next section, got %q", got[2].Body)
	}
	for _, sec := range got {
		if sec.Kind != exchange.KindSection {
			t.Errorf("doc section kind = %q, want section", sec.Kind)
		}
		if len(sec.Fields) > 0 {
			t.Errorf("doc section must carry no fields, got %v", sec.Fields)
		}
	}

	if empty := docSections("   \n"); empty != nil {
		t.Errorf("an empty doc comment yields no sections, got %+v", empty)
	}
}

// TestRunExtract_DocCommentBecomesSection verifies the wiring end to end: the
// symbol keeps the reqmd block and its doc prose moves into a child section, and
// a Markdown heading in the doc does not break the rendered nesting.
func TestRunExtract_DocCommentBecomesSection(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// Foo does foo.
//
// ## Behaviour
//
// It is careful.
func Foo() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")

	var foo exchange.Item
	for _, it := range allItems(alpha.Items) {
		if it.Title == "Foo" {
			foo = it
		}
	}
	if foo.Key == "" {
		t.Fatal("Foo requirement missing")
	}
	if foo.Body != "" {
		t.Errorf("doc prose must move out of the requirement body, got %q", foo.Body)
	}
	if len(foo.Children) != 2 || foo.Children[0].Title != "" || foo.Children[1].Title != "Behaviour" {
		t.Fatalf("Foo children = %+v, want intro section + Behaviour", foo.Children)
	}

	// The rendered document is valid reqmd input. Heading nesting is asserted on
	// the item tree above; how it renders is the renderer's contract and is
	// covered in reqmd-import/exchange, so it is not re-asserted as text here.
	files := renderTo(t, bundle)
	if _, ok := files["alpha/package.md"]; !ok {
		t.Fatalf("missing alpha/package.md; got %v", keys(files))
	}
}

// TestDocSections_GodocSeparator guards against a decorative rule becoming a
// setext heading: "-----" under a text line renders as an h2, which would turn
// a Godoc separator into a section that swallows the symbols after it.
func TestDocSections_GodocSeparator(t *testing.T) {
	doc := "First paragraph.\n\n" +
		"---------------------------------------------------------------------------\n" +
		" x-reqmd.requires-trace-from document default (SYS-FMT-003)\n" +
		" ---------------------------------------------------------------------------\n\n" +
		"Second paragraph.\n"
	got := docSections(doc)
	if len(got) != 1 {
		t.Fatalf("want a single intro section, got %d: %+v", len(got), got)
	}
	body := got[0].Body
	for _, banned := range []string{"-----", "------"} {
		if strings.Contains(body, banned) {
			t.Errorf("decorative rule leaked into the prose: %q", body)
		}
	}
	if !strings.Contains(body, "First paragraph.") || !strings.Contains(body, "Second paragraph.") {
		t.Errorf("prose around the separator was lost: %q", body)
	}
	if !strings.Contains(body, "x-reqmd.requires-trace-from document default") {
		t.Errorf("the separator's caption should survive as prose: %q", body)
	}
}

// runTestExtractErr runs an extraction that is expected to fail.
func runTestExtractErr(t *testing.T, args []string, f *extractFlags) (string, string, error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	err := runExtract(cmd, args, f)
	return outBuf.String(), errBuf.String(), err
}

// itemByTitle finds one item in a document tree by its title.
func itemByTitle(t *testing.T, items []exchange.Item, title string) exchange.Item {
	t.Helper()
	for _, it := range allItems(items) {
		if it.Title == title {
			return it
		}
	}
	t.Fatalf("item %q not found", title)
	return exchange.Item{}
}

// TestRunExtract_ExplicitSymbolLink covers the explicit marker path: a test
// traces to a same-package function by name, and the link survives rendering.
func TestRunExtract_ExplicitSymbolLink(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// Parse parses.
func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

// reqmd:trace Parse
func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	parse := itemByTitle(t, alpha.Items, "Parse")
	testParse := itemByTitle(t, alpha.Items, "TestParse")
	if !slices.Equal(testParse.Links, []string{parse.Key}) {
		t.Errorf("TestParse links = %v, want [%s]", testParse.Links, parse.Key)
	}
	files := renderTo(t, bundle)
	md := files["alpha/package.md"]
	if !strings.Contains(md, "trace:\n") || !strings.Contains(md, "IMP-alpha-"+parse.Key) {
		t.Errorf("rendered trace missing %q:\n%s", "IMP-alpha-"+parse.Key, md)
	}
}

// TestRunExtract_MethodScopes covers receiver-qualified references and the
// C++ `::` spelling.
func TestRunExtract_MethodScopes(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

type Parser struct{}

// Parse parses.
func (p *Parser) Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

// reqmd:trace Parser.Parse
func TestDotted() {}

// reqmd:trace Parser::Parse
func TestScoped() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	method := itemByTitle(t, alpha.Items, "Parse")
	for _, name := range []string{"TestDotted", "TestScoped"} {
		it := itemByTitle(t, alpha.Items, name)
		if !slices.Equal(it.Links, []string{method.Key}) {
			t.Errorf("%s links = %v, want [%s]", name, it.Links, method.Key)
		}
	}
}

// TestRunExtract_BasePackageResolution covers external test packages, where
// the test lives in a different document from the symbol it tests.
func TestRunExtract_BasePackageResolution(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// Parse parses.
func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha_test

// reqmd:trace Parse
func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	parse := itemByTitle(t, alpha.Items, "Parse")
	ext := docByDir(t, bundle, "alpha_test")
	testParse := itemByTitle(t, ext.Items, "TestParse")
	want := "alpha/" + parse.Key
	if !slices.Equal(testParse.Links, []string{want}) {
		t.Errorf("links = %v, want [%s]", testParse.Links, want)
	}
}

// TestRunExtract_UnresolvedSymbolIsFatal pins the hard-error contract: no
// bundle, no stdout, and an existing -o file is left untouched.
func TestRunExtract_UnresolvedSymbolIsFatal(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a_test.go", `package alpha

// reqmd:trace dummy
func TestDummy() {}
`)
	out, errOut, err := runTestExtractErr(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	if err == nil {
		t.Fatal("expected an error for an unresolved symbol")
	}
	if !strings.Contains(err.Error(), `unresolved symbol trace "dummy"`) {
		t.Errorf("error = %v, want unresolved symbol message", err)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty for a data error", errOut)
	}

	// -o must not be created or truncated on failure.
	dst := filepath.Join(t.TempDir(), "bundle.json")
	if err := os.WriteFile(dst, []byte("sentinel"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTestExtractErr(t, []string{src}, &extractFlags{idPrefix: "IMP-", output: dst}); err == nil {
		t.Fatal("expected an error when writing to -o")
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "sentinel" {
		t.Errorf("-o file = %q, want it untouched", got)
	}
}

// TestRunExtract_AmbiguousSymbolIsFatal covers a reference that matches the
// same name in two packages.
func TestRunExtract_AmbiguousSymbolIsFatal(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "alpha/a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "beta/b.go", `package beta

func Parse() {}
`)
	writeFile(t, src, "gamma/c.go", `package gamma

// reqmd:trace Parse
func TestParse() {}
`)
	_, _, err := runTestExtractErr(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	if err == nil {
		t.Fatal("expected an ambiguity error")
	}
	for _, want := range []string{"ambiguous symbol trace", "alpha/Parse", "beta/Parse"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want %q", err, want)
		}
	}
}

// TestRunExtract_MalformedMarkerIsFatal covers prose left on the directive.
func TestRunExtract_MalformedMarkerIsFatal(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// reqmd:trace see the parser
func Foo() {}
`)
	_, _, err := runTestExtractErr(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	if err == nil {
		t.Fatal("expected a malformed marker error")
	}
	if !strings.Contains(err.Error(), "invalid reqmd:trace marker") {
		t.Errorf("error = %v, want malformed marker message", err)
	}
}

// TestRunExtract_QualifiedAndOptOut covers an explicit package qualifier and
// the `none` opt-out, which must not create a link or an error.
func TestRunExtract_QualifiedAndOptOut(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "alpha/a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "gamma/c.go", `package gamma

// reqmd:trace alpha/Parse
func TestQualified() {}

// reqmd:trace none
func TestOptOut() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	parse := itemByTitle(t, alpha.Items, "Parse")
	gamma := docByDir(t, bundle, "gamma")
	qualified := itemByTitle(t, gamma.Items, "TestQualified")
	if !slices.Equal(qualified.Links, []string{"alpha/" + parse.Key}) {
		t.Errorf("qualified links = %v, want [alpha/%s]", qualified.Links, parse.Key)
	}
	optOut := itemByTitle(t, gamma.Items, "TestOptOut")
	if len(optOut.Links) != 0 {
		t.Errorf("opt-out item links = %v, want none", optOut.Links)
	}
}

// TestRunExtract_PythonTestLinksToModule covers a test file in its own
// document reaching back to the module it tests.
func TestRunExtract_PythonTestLinksToModule(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "parse.py", `def parse():
    pass
`)
	writeFile(t, src, "test_parse.py", `#reqmd:trace parse
def test_parse():
    pass
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	parse := docByDir(t, bundle, "parse")
	fn := itemByTitle(t, parse.Items, "parse")
	tests := docByDir(t, bundle, "test_parse")
	testFn := itemByTitle(t, tests.Items, "test_parse")
	want := "parse/" + fn.Key
	if !slices.Equal(testFn.Links, []string{want}) {
		t.Errorf("test_parse links = %v, want [%s]", testFn.Links, want)
	}
}

// TestRunExtract_RequirementIDRefsPassThrough keeps the ID path intact: an
// ID-shaped marker becomes a ref, not a symbol lookup, and renders into trace.
func TestRunExtract_RequirementIDRefsPassThrough(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

// reqmd:trace SW-EXP-001
func Parse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	parse := itemByTitle(t, alpha.Items, "Parse")
	if !slices.Equal(parse.Refs, []string{"SW-EXP-001"}) {
		t.Errorf("Refs = %v, want [SW-EXP-001]", parse.Refs)
	}
	if len(parse.Links) != 0 {
		t.Errorf("Links = %v, want none for an ID ref", parse.Links)
	}
	if md := renderTo(t, bundle)["alpha/package.md"]; !strings.Contains(md, "SW-EXP-001") {
		t.Errorf("rendered trace missing the ref:\n%s", md)
	}
}

// TestRunExtract_CaseMismatchIsFatal pins that explicit matching is
// case-sensitive: a near-miss is an error, not a silent skip.
func TestRunExtract_CaseMismatchIsFatal(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

// reqmd:trace parse
func TestParse() {}
`)
	_, _, err := runTestExtractErr(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	if err == nil || !strings.Contains(err.Error(), `unresolved symbol trace "parse"`) {
		t.Fatalf("err = %v, want an unresolved symbol error", err)
	}
}

// TestRunExtract_BuildTagDuplicateNotAmbiguous pins that build-tag twins are
// deduplicated before indexing, so a reference to them resolves once.
func TestRunExtract_BuildTagDuplicateNotAmbiguous(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "graph.go", `package export

func New() {}
`)
	writeFile(t, src, "graph_stub.go", `package export

func New() {}
`)
	writeFile(t, src, "graph_test.go", `package export_test

// reqmd:trace New
func TestNew() {}
`)
	bundle, _, errOut := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	export := docByDir(t, bundle, "export")
	target := itemByTitle(t, export.Items, "New")
	ext := docByDir(t, bundle, "export_test")
	testNew := itemByTitle(t, ext.Items, "TestNew")
	if !slices.Equal(testNew.Links, []string{"export/" + target.Key}) {
		t.Errorf("links = %v, want [export/%s]", testNew.Links, target.Key)
	}
	if !strings.Contains(errOut, "skipping duplicate symbol New") {
		t.Errorf("stderr = %q, want the duplicate-symbol warning", errOut)
	}
}

// TestRunExtract_DuplicateLinksDeduped pins that two spellings of the same
// target produce one link.
func TestRunExtract_DuplicateLinksDeduped(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

// reqmd:trace Parse, alpha/Parse
func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	parse := itemByTitle(t, alpha.Items, "Parse")
	testParse := itemByTitle(t, alpha.Items, "TestParse")
	if !slices.Equal(testParse.Links, []string{parse.Key}) {
		t.Errorf("links = %v, want one [%s]", testParse.Links, parse.Key)
	}
}

// TestFormatTraceRefErrors pins deterministic sort order and the cap.
func TestFormatTraceRefErrors(t *testing.T) {
	var errs []traceRefError
	for i := 0; i < 22; i++ {
		errs = append(errs, traceRefError{
			file:    "z.go",
			line:    100 - i,
			symbol:  "TestX",
			message: "unresolved symbol trace \"X\"",
		})
	}
	err := formatTraceRefErrors(errs)
	if err == nil {
		t.Fatal("want an error")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "invalid reqmd:trace references:") {
		t.Errorf("message = %q", msg)
	}
	if !strings.Contains(msg, "z.go:79 (TestX)") || !strings.Contains(msg, "z.go:98 (TestX)") {
		t.Errorf("entries are not ascending by line:\n%s", msg)
	}
	if strings.Contains(msg, "z.go:99 (TestX)") {
		t.Errorf("cap not applied; 21st entry was printed:\n%s", msg)
	}
	if !strings.Contains(msg, "... and 2 more") {
		t.Errorf("missing cap notice:\n%s", msg)
	}
}

// testLinks returns an item's links for assertions.
func testLinks(t *testing.T, items []exchange.Item, title string) []string {
	t.Helper()
	return itemByTitle(t, items, title).Links
}

// TestRunExtract_LinkTestsOffByDefault pins that conventions are opt-in.
func TestRunExtract_LinkTestsOffByDefault(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	if links := testLinks(t, alpha.Items, "TestParse"); len(links) != 0 {
		t.Errorf("links = %v, want none without --link-tests", links)
	}
}

// TestRunExtract_ConventionGo covers CamelCase matching: longest candidate,
// CamelCase boundaries, initial-case folding, methods, types, and the target
// pool exclusions.
func TestRunExtract_ConventionGo(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}
func parse() {}
func ParseFromDisk() {}
func URLParser() {}
func cleanComment() {}
type Point struct{}
type Graph struct{}

func (g *Graph) CheckResults() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

func TestParse() {}
func TestParseFromDisk() {}
func TestURLParser() {}
func TestCleanComment() {}
func TestPoint() {}
func TestGraph_CheckResults() {}
func TestGraph_CheckResults_Empty() {}
func TestParseHelper() {}
func TestMain(m *testing.M) {}

type ParseHelper struct{}
`)
	bundle, _, errOut := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	alpha := docByDir(t, bundle, "alpha")
	key := func(title string) string { return itemByTitle(t, alpha.Items, title).Key }

	cases := map[string]string{
		"TestParse":                    key("Parse"),
		"TestParseFromDisk":            key("ParseFromDisk"),
		"TestURLParser":                key("URLParser"),
		"TestCleanComment":             key("cleanComment"),
		"TestPoint":                    key("Point"),
		"TestGraph_CheckResults":       key("CheckResults"),
		"TestGraph_CheckResults_Empty": key("CheckResults"),
		// The test-file type ParseHelper must not win over the production Parse.
		"TestParseHelper": key("Parse"),
	}
	for test, want := range cases {
		if got := testLinks(t, alpha.Items, test); !slices.Equal(got, []string{want}) {
			t.Errorf("%s links = %v, want [%s]", test, got, want)
		}
	}
	if links := testLinks(t, alpha.Items, "TestMain"); len(links) != 0 {
		t.Errorf("TestMain links = %v, want none", links)
	}
	// Exact case outranks the folded lowercase candidate `parse`.
	if got := testLinks(t, alpha.Items, "TestParse"); !slices.Equal(got, []string{key("Parse")}) {
		t.Errorf("TestParse links = %v, want the exact-case Parse", got)
	}
	if strings.Contains(errOut, "ambiguous") {
		t.Errorf("unexpected ambiguity warning: %s", errOut)
	}
}

// TestRunExtract_ConventionGoExternalTestPackage covers the `_test` half of
// basePackage: a test in package alpha_test links to package alpha.
func TestRunExtract_ConventionGoExternalTestPackage(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha_test

func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	alpha := docByDir(t, bundle, "alpha")
	parse := itemByTitle(t, alpha.Items, "Parse")
	ext := docByDir(t, bundle, "alpha_test")
	if got := testLinks(t, ext.Items, "TestParse"); !slices.Equal(got, []string{"alpha/" + parse.Key}) {
		t.Errorf("links = %v, want [alpha/%s]", got, parse.Key)
	}
}

// TestRunExtract_ConventionPython covers snake_case across test documents.
func TestRunExtract_ConventionPython(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "parse.py", `def parse():
    pass
`)
	writeFile(t, src, "test_parse.py", `def test_parse():
    pass
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	parse := docByDir(t, bundle, "parse")
	fn := itemByTitle(t, parse.Items, "parse")
	tests := docByDir(t, bundle, "test_parse")
	if got := testLinks(t, tests.Items, "test_parse"); !slices.Equal(got, []string{"parse/" + fn.Key}) {
		t.Errorf("links = %v, want [parse/%s]", got, fn.Key)
	}
}

// TestRunExtract_ConventionRustInline covers an inline #[test] module.
func TestRunExtract_ConventionRustInline(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.rs", `fn parse() {}

mod tests {
    #[test]
    fn test_parse() {}
}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	a := docByDir(t, bundle, "a")
	fn := itemByTitle(t, a.Items, "parse")
	if got := testLinks(t, a.Items, "test_parse"); !slices.Equal(got, []string{fn.Key}) {
		t.Errorf("links = %v, want [%s]", got, fn.Key)
	}
}

// TestRunExtract_ConventionExplicitWins pins that any explicit directive
// (an ID ref or none) suppresses convention linking.
func TestRunExtract_ConventionExplicitWins(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

// reqmd:trace SW-EXP-001
func TestParse() {}

// reqmd:trace none
func TestParseB() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	alpha := docByDir(t, bundle, "alpha")
	if links := testLinks(t, alpha.Items, "TestParse"); len(links) != 0 {
		t.Errorf("explicit ID should suppress convention, links = %v", links)
	}
	if refs := itemByTitle(t, alpha.Items, "TestParse").Refs; !slices.Equal(refs, []string{"SW-EXP-001"}) {
		t.Errorf("refs = %v, want [SW-EXP-001]", refs)
	}
	if links := testLinks(t, alpha.Items, "TestParseB"); len(links) != 0 {
		t.Errorf("none should suppress convention, links = %v", links)
	}
}

// TestRunExtract_ConventionAmbiguousWarns pins non-fatal tie handling.
func TestRunExtract_ConventionAmbiguousWarns(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

type A_B struct{}
type A struct{}

func (x *A_B) C() {}
func (x *A) B_C() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

func TestA_B_C() {}
`)
	bundle, _, errOut := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	alpha := docByDir(t, bundle, "alpha")
	if links := testLinks(t, alpha.Items, "TestA_B_C"); len(links) != 0 {
		t.Errorf("ambiguous match produced links %v", links)
	}
	if !strings.Contains(errOut, "ambiguous convention match") {
		t.Errorf("stderr = %q, want an ambiguity warning", errOut)
	}
}

// TestRunExtract_ConventionNoGlobalFallback pins that a test in one package
// cannot link to a same-named symbol in an unrelated package.
func TestRunExtract_ConventionNoGlobalFallback(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "alpha/a.go", `package alpha

func Parse() {}
`)
	writeFile(t, src, "beta/b_test.go", `package beta

func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	beta := docByDir(t, bundle, "beta")
	if links := testLinks(t, beta.Items, "TestParse"); len(links) != 0 {
		t.Errorf("cross-package convention link = %v, want none", links)
	}
}

// TestRunExtract_ConventionMethodBareName pins the fallback that links a test
// to a method when the test names the method without its receiver.
func TestRunExtract_ConventionMethodBareName(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package clang

type CLanguage struct{}

func (l *CLanguage) Parse() {}
`)
	writeFile(t, src, "a_test.go", `package clang

func TestParseSample() {}
`)
	bundle, _, errOut := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	clang := docByDir(t, bundle, "clang")
	parse := itemByTitle(t, clang.Items, "Parse")
	if got := testLinks(t, clang.Items, "TestParseSample"); !slices.Equal(got, []string{parse.Key}) {
		t.Errorf("links = %v, want the Parse method [%s]", got, parse.Key)
	}
	if strings.Contains(errOut, "ambiguous") {
		t.Errorf("unexpected ambiguity: %s", errOut)
	}
}

// TestRunExtract_ConventionFunctionBeatsBareMethod pins the ranking: a real
// function wins over a bare method name at the same match length.
func TestRunExtract_ConventionFunctionBeatsBareMethod(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

func Parse() {}

type Parser struct{}

func (p *Parser) Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

func TestParse() {}
`)
	bundle, _, errOut := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	alpha := docByDir(t, bundle, "alpha")
	fn := itemByTitle(t, alpha.Items, "Parse")
	if got := testLinks(t, alpha.Items, "TestParse"); !slices.Equal(got, []string{fn.Key}) {
		t.Errorf("links = %v, want the top-level Parse [%s]", got, fn.Key)
	}
	if strings.Contains(errOut, "ambiguous") {
		t.Errorf("unexpected ambiguity: %s", errOut)
	}
}

// TestRunExtract_ConventionTwoBareMethodsAmbiguous pins the tie: two methods
// with the same bare name cannot be chosen between.
func TestRunExtract_ConventionTwoBareMethodsAmbiguous(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

type A struct{}
type B struct{}

func (x *A) Parse() {}
func (x *B) Parse() {}
`)
	writeFile(t, src, "a_test.go", `package alpha

func TestParseSample() {}
`)
	bundle, _, errOut := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-", linkTests: true})
	alpha := docByDir(t, bundle, "alpha")
	if links := testLinks(t, alpha.Items, "TestParseSample"); len(links) != 0 {
		t.Errorf("ambiguous bare methods produced links %v", links)
	}
	if !strings.Contains(errOut, "ambiguous convention match") {
		t.Errorf("stderr = %q, want an ambiguity warning", errOut)
	}
}

// TestRunExtract_MethodInOtherFileKeepsOwnLevel pins that a method whose type
// is declared in another file is still emitted (at its own file's level)
// instead of being dropped during nesting.
func TestRunExtract_MethodInOtherFileKeepsOwnLevel(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.go", `package alpha

type Parser struct{}
`)
	writeFile(t, src, "b.go", `package alpha

// Parse parses.
func (p *Parser) Parse() {}
`)
	writeFile(t, src, "b_test.go", `package alpha

// reqmd:trace Parser.Parse
func TestParse() {}
`)
	bundle, _, _ := runTestExtract(t, []string{src}, &extractFlags{idPrefix: "IMP-"})
	alpha := docByDir(t, bundle, "alpha")
	method := itemByTitle(t, alpha.Items, "Parse")
	if method.Key == "" {
		t.Fatal("method in another file was dropped")
	}
	if !strings.Contains(method.Fields["x-reqmd.source-file"].(string), "b.go") {
		t.Errorf("method provenance = %v, want b.go", method.Fields["x-reqmd.source-file"])
	}
	test := itemByTitle(t, alpha.Items, "TestParse")
	if !slices.Equal(test.Links, []string{method.Key}) {
		t.Errorf("links = %v, want [%s]", test.Links, method.Key)
	}
}
