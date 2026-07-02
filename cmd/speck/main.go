// Command speck turns a rough idea into a structured SPEC.md.
package main

import (
	"os"

	"github.com/palladius/speck/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
