package main

import (
	"github.com/spf13/cobra"
)

// Version is injected by the linker at build time via
// -ldflags "-X main.Version=<value>". Falls back to "dev" for local builds.
var Version = "dev"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "skillhire",
		Short:         "Recruit, compose, and deploy reusable AI skills.",
		Long:          "Skills for Hire is a versioned registry and toolkit for distributing autonomous AI skills.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newValidateCmd())
	root.AddCommand(newPackCmd())
	root.AddCommand(newLockCmd())
	root.AddCommand(newVersionCmd())
	return root
}
