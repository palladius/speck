// Command speck turns a rough idea into a structured SPEC.md.
package main

import (
	"os"

	"github.com/palladius/speck/internal/cli"
	"github.com/palladius/speck/internal/dotenv"
)

func main() {
	_ = dotenv.Load(".env")
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
