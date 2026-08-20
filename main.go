package main

import (
	"fmt"
	"os"

	"github.com/r3drun3/prwlrctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
