package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const bundleJSON = `{
  "format": "reqmd-import/v1",
  "documents": [
    {"id-prefix": "IMP-", "level": "imported-requirements"},
    {"dir": "alpha", "file": "package.md", "title": "Package: alpha", "id-prefix": "IMP-alpha-",
     "requirements": [{"native-id": "Foo-abc1234", "title": "Foo", "attrs": {"x-reqmd.imported": true}}]}
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

func TestRunRejectsBadBundle(t *testing.T) {
	out := t.TempDir()
	cmd := NewRootCmd("test")
	cmd.SetIn(strings.NewReader(`{"format":"nope","documents":[]}`))
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-", "-o", out})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error for an unsupported format")
	}
}
