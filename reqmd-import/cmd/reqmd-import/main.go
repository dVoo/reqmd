// Command reqmd-import renders a reqmd-import/v1 exchange bundle into reqmd
// documents (schema.yaml + .md) that the reqmd CLI can validate.
package main

import (
	"fmt"
	"os"

	"github.com/dVoo/reqmd/reqmd-import/internal/cli"
)

// version is overridden at release time via:
//
//	go build -ldflags "-X main.version=v0.1.0"
var version string

func main() {
	root := cli.NewRootCmd(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
