package main

import (
	"os"

	"github.com/mmdbasi/mtcli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
