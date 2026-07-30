package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"

	"reqmd/internal/cli"
	"reqmd/internal/reporter"
)

func main() {
	// GOGC tuning: allow users to set a higher GC percent via the
	// REQMD_GOGC env var. Default is Go's 100. Higher values (e.g. 200)
	// reduce GC frequency at the cost of higher peak RSS.
	if v := os.Getenv("REQMD_GOGC"); v != "" {
		if pct, err := strconv.Atoi(v); err == nil {
			debug.SetGCPercent(pct)
		}
	}

	root := cli.NewRootCmd()
	err := root.Execute()
	if err == nil {
		return
	}
	var exitErr *reporter.ExitCodeError
	if errors.As(err, &exitErr) {
		fmt.Fprintln(os.Stderr, exitErr)
		os.Exit(exitErr.Code)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
