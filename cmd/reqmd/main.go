// Command reqmd validates and exports requirement specifications stored in
// Markdown files with embedded attr blocks.
package main

import (
	"errors"
	"fmt"
	"os"
	"reqmd/internal/cli"
	"runtime/debug"
	"strconv"
)

// exitCoder is implemented by errors that carry a process exit code
// (reporter.ExitCodeError). main depends on the behavior, not the
// concrete type, so the entry point stays decoupled from the reporter
// package.
type exitCoder interface {
	ExitCode() int
}

// version is overridden at release time via:
//
//	go build -ldflags "-X main.version=v0.1.0"
//
// When empty (dev builds), `reqmd --version` reports "(devel)".
var version string

func main() {
	// GOGC tuning: allow users to set a higher GC percent via the
	// REQMD_GOGC env var. Default is Go's 100. Higher values (e.g. 200)
	// reduce GC frequency at the cost of higher peak RSS.
	if v := os.Getenv("REQMD_GOGC"); v != "" {
		pct, err := strconv.Atoi(v)
		if err == nil {
			debug.SetGCPercent(pct)
		}
	}

	root := cli.NewRootCmd(version)

	err := root.Execute()
	if err == nil {
		return
	}

	var exitErr exitCoder

	if errors.As(err, &exitErr) {
		fmt.Fprintln(os.Stderr, exitErr)
		os.Exit(exitErr.ExitCode())
	}

	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
