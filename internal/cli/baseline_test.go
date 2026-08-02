package cli

import (
	"strings"
	"testing"
)

// TestBaselineDiff_FlagGroupValidation verifies cobra flag-group rules:
// --filter-a/--filter-b are required together (RFC §3.4) and --filter is
// mutually exclusive with the two-view mode. These are validated before
// RunE, so no git repo is required.
func TestBaselineDiff_FlagGroupValidation(t *testing.T) {
	t.Run("filter-a requires filter-b", func(t *testing.T) {
		_, err := runCmd(t, newBaselineDiffCmd(), "--filter-a", `"Base" in variant`)
		if err == nil || !strings.Contains(err.Error(), "must all be set") {
			t.Fatalf("err = %v, want 'must all be set'", err)
		}
	})
	t.Run("filter-b requires filter-a", func(t *testing.T) {
		_, err := runCmd(t, newBaselineDiffCmd(), "--filter-b", `"Premium" in variant`)
		if err == nil || !strings.Contains(err.Error(), "must all be set") {
			t.Fatalf("err = %v, want 'must all be set'", err)
		}
	})
	t.Run("filter mutually exclusive with views", func(t *testing.T) {
		_, err := runCmd(t, newBaselineDiffCmd(),
			"--filter", `"Base" in variant`,
			"--filter-a", `"Base" in variant`,
			"--filter-b", `"Premium" in variant`)
		if err == nil || !strings.Contains(err.Error(), "none of the others can be") {
			t.Fatalf("err = %v, want mutually-exclusive error", err)
		}
	})
}

// TestBaselineDiff_TwoTagFilter compares two tags and verifies --filter
// scopes both snapshots (RFC §3.3).
func TestBaselineDiff_TwoTagFilter(t *testing.T) {
	repo := t.TempDir()
	git := gitHelper(t, repo)
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")

	base := "# Software\n\n" +
		mdReq("SW-001", "Base parser", "status: approved\nvariant: [Base]\n") +
		mdReq("SW-002", "Premium schema", "status: approved\nvariant: [Premium]\n")
	writeFile(t, repo, "sw/schema.yaml", variantSchema)
	writeFile(t, repo, "sw/sw.md", base)
	git("add", "-A")
	git("commit", "-qm", "v1")
	git("tag", "v1")

	writeFile(t, repo, "sw/sw.md", base+mdReq("SW-003", "Premium export", "status: approved\nvariant: [Premium]\n"))
	git("add", "-A")
	git("commit", "-qm", "v2")
	git("tag", "v2")

	t.Chdir(repo)

	// Unfiltered: SW-003 appears as added.
	out, err := runCmd(t, newBaselineDiffCmd(), "v1", "v2")
	if err != nil {
		t.Fatalf("baseline diff failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "SW-003") {
		t.Fatalf("unfiltered diff should show SW-003 added:\n%s", out)
	}

	// Filter to Base: SW-003 (Premium) is excluded from both snapshots.
	out, err = runCmd(t, newBaselineDiffCmd(), "v1", "v2", "--filter", `"Base" in variant`)
	if err != nil {
		t.Fatalf("baseline diff --filter failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "SW-003") {
		t.Fatalf("filtered diff should not show SW-003:\n%s", out)
	}
}

// TestBaselineDiff_FilteredViews verifies the same-commit two-view mode
// (RFC §3.4): "what Premium adds over Base" from a single commit.
func TestBaselineDiff_FilteredViews(t *testing.T) {
	repo := t.TempDir()
	git := gitHelper(t, repo)
	git("init", "-q")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")

	writeFile(t, repo, "sw/schema.yaml", variantSchema)
	writeFile(t, repo, "sw/sw.md", "# Software\n\n"+
		mdReq("SW-001", "Base parser", "status: approved\nvariant: [Base]\n")+
		mdReq("SW-002", "Premium schema", "status: approved\nvariant: [Premium]\n"))
	git("add", "-A")
	git("commit", "-qm", "base")

	t.Chdir(repo)

	out, err := runCmd(t, newBaselineDiffCmd(),
		"--filter-a", `"Base" in variant`,
		"--filter-b", `"Premium" in variant`)
	if err != nil {
		t.Fatalf("baseline diff --filter-a/--filter-b failed: %v\n%s", err, out)
	}
	// View A (Base) holds SW-001; view B (Premium) holds SW-002. The diff
	// reports SW-001 removed and SW-002 added.
	if !strings.Contains(out, "SW-002") {
		t.Fatalf("filtered-views diff should show SW-002 added:\n%s", out)
	}
	if !strings.Contains(out, "SW-001") {
		t.Fatalf("filtered-views diff should show SW-001 removed:\n%s", out)
	}
}
