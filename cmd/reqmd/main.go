package main

import (
	"errors"
	"fmt"
	"os"

	"reqmd/internal/cli"
	"reqmd/internal/reporter"
)

func main() {
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
