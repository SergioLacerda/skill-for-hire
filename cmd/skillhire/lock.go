package main

import (
	"fmt"
	"path/filepath"

	"github.com/SergioLacerda/skill-for-hire/internal/lockfile"
	"github.com/spf13/cobra"
)

func newLockCmd() *cobra.Command {
	var (
		outPath    string
		source     string
		verifyOnly bool
	)
	cmd := &cobra.Command{
		Use:   "lock <release.yaml> [<release.yaml>...]",
		Short: "Produce or verify a skillhire.lock from release.yaml sidecars.",
		Long: `Reads one or more <name>-<version>.release.yaml manifests (produced by
'skillhire pack') and writes a deterministic skillhire.lock pinning each
skill to its exact version, digest, size, and source.

Each release.yaml is treated as one skill. --source defaults to
"monorepo:<basename>-without-extension" and can be overridden per invocation.
When --verify is set, the command compares the freshly built lockfile
against the file at --out and exits non-zero on any drift.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			inputs := make([]lockfile.ManifestInput, 0, len(args))
			for _, path := range args {
				inputs = append(inputs, lockfile.ManifestInput{
					Path:         path,
					Source:       source, // empty defaults to monorepo:skills/<name> inside FromManifests
					ResolvedFrom: filepath.Base(path),
				})
			}

			lf, err := lockfile.FromManifests(inputs, "skillhire@"+Version)
			if err != nil {
				return err
			}

			body, err := lockfile.Render(lf)
			if err != nil {
				return err
			}

			if verifyOnly {
				existing, err := lockfile.Load(outPath)
				if err != nil {
					return fmt.Errorf("lock --verify: %w", err)
				}
				// Compare only the pinned state (installed map). The
				// generator field is informational metadata and naturally
				// varies across local dev, CI, and tagged builds, so it
				// must not fail a verify.
				if diff := lockfile.DiffInstalled(existing, lf); diff != "" {
					return fmt.Errorf("lock --verify: %s is out of date; regenerate with 'skillhire lock'\n%s", outPath, diff)
				}
				if _, err := fmt.Fprintf(out, "verified %s (%d skill(s))\n", outPath, len(lf.Installed)); err != nil {
					return fmt.Errorf("write verify summary: %w", err)
				}
				return nil
			}

			// `body` is available for the future --stdout mode; drop it
			// on the floor here to keep --verify path minimal.
			_ = body

			if err := lockfile.Write(outPath, lf); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "wrote %s (%d skill(s))\n", outPath, len(lf.Installed)); err != nil {
				return fmt.Errorf("write lock summary: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&outPath, "out", "skillhire.lock", "Output lockfile path.")
	cmd.Flags().StringVar(&source, "source", "", "Override source string for every input (default is monorepo:skills/<name> from the manifest).")
	cmd.Flags().BoolVar(&verifyOnly, "verify", false, "Do not write; fail if --out differs from the freshly built lockfile.")
	return cmd
}
