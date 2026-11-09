// Command refweir provides the command-line interface for the refweir.dev Git forge.
// It supports managing repositories, pull requests, applications, and authentication.
package main

import (
	"fmt"
	"os"

	"refweir.dev/internal/cli"
)

var version = "dev"

func main() {
	cli.Version = version
	c := cli.New()
	if err := c.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "refweir: %v\n", err)
		os.Exit(1)
	}
}
