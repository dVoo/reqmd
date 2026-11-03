package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const schemaURL = "https://reqmd.dev/exchange/v1/schema.json"

const bundleJSON = `{
  "$schema": "` + schemaURL + `",
  "documents": [
    {"key": "root", "title": "Imported proxies", "x-reqmd": {"id-prefix": "IMP-", "level": "imported-requirements", "dir": "."}},
    {"key": "alpha", "title": "Package: alpha",
     "x-reqmd": {"id-prefix": "IMP-alpha-", "dir": "alpha", "file": "package.md"},
     "items": [{"kind": "requirement", "key": "Foo-abc1234", "title": "Foo",
                "fields": {"x-reqmd.imported": true}}]}
  ]
}`

func TestRunFromFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.json")
	if err := os.WriteFile(src, []byte(bundleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")

	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test")
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{src, "-o", out})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v (stderr: %s)", err, errBuf.String())
	}

	for _, want := range []string{"schema.yaml", "alpha/schema.yaml", "alpha/package.md"} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("missing output %s: %v", want, err)
		}
	}
	if !strings.Contains(errBuf.String(), "wrote 3 file(s)") {
		t.Errorf("summary should report file count; got %q", errBuf.String())
	}
}

func TestRunFromStdin(t *testing.T) {
	out := t.TempDir()

	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test")
	cmd.SetIn(strings.NewReader(bundleJSON))
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"-", "-o", out})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v (stderr: %s)", err, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(out, "alpha", "package.md")); err != nil {
		t.Errorf("missing alpha/package.md: %v", err)
	}
}

// TestRunValidateOnly covers the conformance path for third-party producers:
// report every violation, write nothing.
func TestRunValidateOnly(t *testing.T) {
	out := t.TempDir()

	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test")
	cmd.SetIn(strings.NewReader(bundleJSON))
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"-", "-o", out, "--validate-only"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v (stderr: %s)", err, errBuf.String())
	}
	if !strings.Contains(outBuf.String(), "bundle is valid") {
		t.Errorf("want a validity report on stdout, got %q", outBuf.String())
	}
	if entries, err := os.ReadDir(out); err == nil && len(entries) > 0 {
		t.Errorf("--validate-only must not write files, found %d", len(entries))
	}

	// A broken bundle reports more than one problem in the same run.
	errCmd := NewRootCmd("test")
	errCmd.SetIn(strings.NewReader(`{"$schema":"` + schemaURL + `","documents":[
	  {"key":"d","title":"D","items":[
	    {"kind":"bogus","title":"a"},
	    {"kind":"requirement","title":"no key"}]}]}`))
	errCmd.SetOut(&bytes.Buffer{})
	errCmd.SetErr(&bytes.Buffer{})
	errCmd.SetArgs([]string{"-", "--validate-only"})
	err := errCmd.Execute()
	if err == nil {
		t.Fatal("expected a validation error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "kind \"bogus\"") || !strings.Contains(msg, "without a key") {
		t.Errorf("violations should be batched, got:\n%s", msg)
	}
}

// TestRunStatusMap covers the renderer-side status mapping flag.
func TestRunStatusMap(t *testing.T) {
	out := t.TempDir()
	b := `{"$schema":"` + schemaURL + `","documents":[{"key":"d","title":"D","items":[
	  {"kind":"requirement","key":"1","title":"R","status":"In Review"}]}]}`

	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd("test")
	cmd.SetIn(strings.NewReader(b))
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"-", "-o", out, "--status-map", "In Review=review"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v (stderr: %s)", err, errBuf.String())
	}
	md, err := os.ReadFile(filepath.Join(out, "d", "requirements.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md), "status: review") {
		t.Errorf("mapped status missing:\n%s", md)
	}
}

func TestRunRejectsBadBundle(t *testing.T) {
	out := t.TempDir()
	cmd := NewRootCmd("test")
	cmd.SetIn(strings.NewReader(`{"$schema":"nope","documents":[]}`))
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-", "-o", out})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error for an unsupported schema URL")
	}
}
