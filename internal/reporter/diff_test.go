package reporter

import (
	"strings"
	"testing"

	"reqmd/internal/diff"
)

func TestFormatDiff_Submodules(t *testing.T) {
	result := &diff.Result{
		Tag1: "v1.0",
		Tag2: "v2.0",
		Submodules: []diff.SubmoduleChange{
			{Path: "vendor/spec-a", OldSHA: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0", NewSHA: "e4f5g6h7i8j9k0l1m2n3o4p5q6r7s8t9u0v1w2x3", Status: "updated"},
			{Path: "vendor/spec-b", OldSHA: "f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6", Status: "removed"},
			{Path: "vendor/spec-c", NewSHA: "j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9", Status: "added"},
		},
	}
	output := FormatDiff(result)
	if !strings.Contains(output, "--- Submodule Changes ---") {
		t.Error("expected submodule section header")
	}
	if !strings.Contains(output, "vendor/spec-a") {
		t.Error("expected vendor/spec-a in output")
	}
	if !strings.Contains(output, "vendor/spec-b") {
		t.Error("expected vendor/spec-b in output")
	}
	if !strings.Contains(output, "vendor/spec-c") {
		t.Error("expected vendor/spec-c in output")
	}
	// Short hashes (7 chars)
	if !strings.Contains(output, "a1b2c3d") {
		t.Error("expected short hash a1b2c3d")
	}
}

func TestFormatDiff_NoSubmodules(t *testing.T) {
	result := &diff.Result{
		Tag1: "v1.0",
		Tag2: "v2.0",
		// No submodules
	}
	output := FormatDiff(result)
	if strings.Contains(output, "--- Submodule Changes ---") {
		t.Error("expected no submodule section when empty")
	}
}

func TestShortHash(t *testing.T) {
	tests := []struct {
		sha  string
		want string
	}{
		{"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0", "a1b2c3d"},
		{"abc123", "abc123"},
		{"", ""},
		{"1234567", "1234567"},
		{"12345678", "1234567"},
	}
	for _, tt := range tests {
		t.Run(tt.sha, func(t *testing.T) {
			if got := shortHash(tt.sha); got != tt.want {
				t.Errorf("shortHash(%q) = %q, want %q", tt.sha, got, tt.want)
			}
		})
	}
}
