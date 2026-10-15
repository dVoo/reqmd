// Command reqmd-code extracts source-code symbols into a reqmd-import/v1
// exchange bundle, which reqmd-import renders into reqmd requirement files.
package main

import (
	"fmt"
	"os"
	"reqmd-code/internal/cli"

	// Blank-import language plugins so their init() functions self-register
	// into the lang registry. Without these, reqmd-code would walk the
	// source tree and skip every file (no plugin claims any extension).
	// Add new plugins here as they land.
	_ "reqmd-code/internal/lang/c"
	_ "reqmd-code/internal/lang/cpp"
	_ "reqmd-code/internal/lang/go"
	_ "reqmd-code/internal/lang/python"
	_ "reqmd-code/internal/lang/rust"
)

// version is overridden at release time via:
//
//	go build -ldflags "-X main.version=v0.1.0"
//
// When empty (dev builds), `reqmd-code --version` reports "(devel)".
var version string

func main() {
	root := cli.NewRootCmd(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
