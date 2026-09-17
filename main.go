package main

import (
	"fmt"
	"os"

	"github.com/moveeeax/tfvars-lint/cmd"
)

func main() {
	code, err := cmd.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tfvars-lint:", err)
	}
	os.Exit(code)
}
