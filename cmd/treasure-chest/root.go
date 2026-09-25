package main

import (
	"github.com/SergioLacerda/skill-for-hire/skills/treasure-chest/runtime"
	"github.com/spf13/cobra"
)

// Version is injected at build time via
// -ldflags "-X main.Version=<value>".
var Version = "dev"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "treasure-chest",
		Short:         "Standalone CLI for the treasure-chest knowledge skill.",
		Long:          "Prepare, search and explain jewels/potions stored under a workspace root, following the Knowledge API v1 contract.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	adapter := runtime.NewAdapter(Version)

	root.AddCommand(newPrepareCmd(adapter))
	root.AddCommand(newSearchCmd(adapter))
	root.AddCommand(newStatusCmd(adapter))
	root.AddCommand(newExplainCmd(adapter))
	root.AddCommand(newVersionCmd())
	return root
}
