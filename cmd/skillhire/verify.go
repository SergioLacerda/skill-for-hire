package main

import (
	"fmt"

	"github.com/SergioLacerda/skill-for-hire/internal/verify"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <archive.tar.gz> [<archive.tar.gz>...]",
		Short: "Verify a downloaded skill pack against its .sha256 and .release.yaml companions.",
		Long: `Reads <archive>.sha256 and (when present) <archive-stem>.release.yaml
sitting next to the archive, recomputes the sha256, and fails on any
mismatch. The .sha256 file may be in coreutils format ("digest  file")
or a bare hex digest. A .release.yaml sidecar is optional; when
present, its digest and size fields are cross-checked against the
archive.

Use this on a downloaded release asset before extracting it.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			var failed int
			for _, path := range args {
				res, err := verify.Verify(path)
				if err != nil {
					failed++
					if _, werr := fmt.Fprintf(out, "FAIL %s\n  %v\n", path, err); werr != nil {
						return fmt.Errorf("write verify report: %w", werr)
					}
					continue
				}
				manifest := "not present"
				if res.ManifestExists {
					manifest = res.ManifestPath
				}
				if _, err := fmt.Fprintf(out,
					"ok  %s\n  sha256:   %s\n  size:     %d\n  manifest: %s\n",
					path, res.SHA256, res.Size, manifest); err != nil {
					return fmt.Errorf("write verify report: %w", err)
				}
			}
			if failed > 0 {
				return fmt.Errorf("%d archive(s) failed verification", failed)
			}
			return nil
		},
	}
}
