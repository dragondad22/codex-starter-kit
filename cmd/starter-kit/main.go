package main

import (
	"os"

	"github.com/dragondad22/codex-starter-kit/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
