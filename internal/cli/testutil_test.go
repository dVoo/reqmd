package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// writeFile writes content to rel under dir, creating parent directories.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runCmd executes a cobra command with args, capturing stdout+stderr.
func runCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// mdReq renders a requirement heading plus its ```attr YAML block.
func mdReq(id, title, attrs string) string {
	return fmt.Sprintf("## %s: %s\n```attr\n%s```\n\n", id, title, attrs)
}

// variantSchema is a minimal schema declaring the custom `variant`
// attribute. `status`, `trace`, and `requires-trace-from` are reqmd
// built-ins injected automatically, so they need no declaration.
const variantSchema = `$schema: "https://json-schema.org/draft/2020-12/schema"
title: "Test Requirements"
type: object
properties:
  variant:
    type: array
    items:
      type: string
      enum: [Base, Premium]
additionalProperties: false
`

// verifySchema adds the custom `verify` attribute used by measures in the
// --filter + --results test.
const verifySchema = `$schema: "https://json-schema.org/draft/2020-12/schema"
title: "Test Requirements"
type: object
properties:
  variant:
    type: array
    items:
      type: string
      enum: [Base, Premium]
  verify:
    type: string
    enum: [Test, Review]
additionalProperties: false
`

// writeVariantSpec creates a single-document spec with SW-001 (Base) and
// SW-002 (Premium). Returns the spec root.
func writeVariantSpec(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "sw/schema.yaml", variantSchema)
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+
		mdReq("SW-001", "Base parser", "status: approved\nvariant: [Base]\n")+
		mdReq("SW-002", "Premium schema", "status: approved\nvariant: [Premium]\n"))
	return root
}

// writeDisjointSpec creates a two-doc spec whose only trace link (SW-001
// [Base] → SYS-001 [Premium]) has zero variant overlap. Returns the root.
func writeDisjointSpec(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+mdReq("SYS-001", "Platform", "status: approved\nvariant: [Premium]\n"))
	writeFile(t, root, "sw/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sw\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+mdReq("SW-001", "Base client", "status: approved\nvariant: [Base]\ntrace: [sys/SYS-001]\n"))
	return root
}

// writeDisjointCleanSpec is like writeDisjointSpec but the trace link
// overlaps (SW-001 [Base, Premium] → SYS-001 [Premium]) and the sw
// schema declares `x-reqmd.disjoint-check: variant`.
func writeDisjointCleanSpec(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+mdReq("SYS-001", "Platform", "status: approved\nvariant: [Premium]\n"))
	writeFile(t, root, "sw/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sw\n  disjoint-check: variant\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+mdReq("SW-001", "Premium client", "status: approved\nvariant: [Base, Premium]\ntrace: [sys/SYS-001]\n"))
	return root
}

// writeCoverageSpec creates a two-doc V-model: SYS-001 (Base) expects
// coverage from the software level; its only provider SW-001 (Premium)
// traces to it. Returns the root.
func writeCoverageSpec(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n  level: system\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+mdReq("SYS-001", "Platform", "status: approved\nvariant: [Base]\nrequires-trace-from: [software]\n"))
	writeFile(t, root, "sw/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sw\n  level: software\n  upstream:\n    level: system\n    sources:\n      - ../sys/\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+mdReq("SW-001", "Premium client", "status: approved\nvariant: [Premium]\ntrace: [sys/SYS-001]\n"))
	return root
}

// writeMeasureSpec creates a single-doc spec with one Base measure
// (verify: Test) that a manual result can trace to.
func writeMeasureSpec(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "sw/schema.yaml", verifySchema+"\nx-reqmd:\n  document-id: sw\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+mdReq("SW-001", "Base measure", "status: approved\nvariant: [Base]\nverify: Test\n"))
	return root
}

// writeManualResult creates a manual-results dir with a single result of
// the given outcome tracing to measureID. Returns the results dir.
func writeManualResult(t *testing.T, measureID, outcome string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "schema.yaml", `$schema: "https://json-schema.org/draft/2020-12/schema"
title: "Manual Results"
type: object
properties:
  outcome:
    type: string
    enum: [pass, fail, skipped, inconclusive]
  verified-at:
    type: string
additionalProperties: false
`)
	writeFile(t, dir, "results.md", "# Results\n\n"+mdReq("RESULT-1", "review", fmt.Sprintf("outcome: %s\nverified-at: 2025-01-15\ntrace: [%s]\n", outcome, measureID)))
	return dir
}

// gitHelper returns a closure that runs git commands in dir with a
// hermetic identity, failing the test on any error.
func gitHelper(t *testing.T, dir string) func(args ...string) {
	t.Helper()
	return func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}
