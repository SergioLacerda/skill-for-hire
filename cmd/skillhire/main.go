// skillhire — CLI for the Skills for Hire ecosystem.
//
// Ships the minimum publish loop: validate a skill package on disk, produce
// a reproducible archive with SHA-256 companion, and print inventory that
// downstream release tooling can consume.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "skillhire:", err)
		os.Exit(1)
	}
}
