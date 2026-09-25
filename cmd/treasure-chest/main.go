// treasure-chest — standalone CLI protocol for the treasure-chest skill.
//
// Satisfies the "standalone is a requirement" rule from the arch doc
// (§3.5 / §9): the skill must be usable without Strategist. This binary
// exposes Prepare / Search / Status / Explain / Refresh directly against
// the local workspace.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "treasure-chest:", err)
		os.Exit(1)
	}
}
