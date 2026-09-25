package main

import (
	"fmt"

	"github.com/SergioLacerda/skill-for-hire/internal/pack"
	"github.com/SergioLacerda/skill-for-hire/internal/skill"
	"github.com/spf13/cobra"
)

func newPackCmd() *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "pack <skill-dir> [<skill-dir>...]",
		Short: "Produce a reproducible <name>-<version>.tar.gz plus .sha256 for each skill.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			for _, dir := range args {
				report, err := skill.Validate(dir)
				if err != nil {
					return fmt.Errorf("%s: %w", dir, err)
				}
				if !report.OK() {
					// Refuse to pack an invalid skill — every failure would
					// otherwise ship as a signed release asset.
					if err := printReport(out, report); err != nil {
						return err
					}
					return fmt.Errorf("%s: package-static validation failed; refusing to pack", dir)
				}
				result, err := pack.Pack(pack.Options{
					SkillDir:  dir,
					SkillName: report.Manifest.Metadata.Name,
					Version:   report.Manifest.Metadata.Version,
					OutDir:    outDir,
				})
				if err != nil {
					return fmt.Errorf("%s: %w", dir, err)
				}
				if _, err := fmt.Fprintf(out, "packed %s (%d bytes)\n  sha256: %s\n  digest: %s\n", result.ArchivePath, result.Size, result.SHA256, result.ChecksumPath); err != nil {
					return fmt.Errorf("write pack summary: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&outDir, "out", "dist", "Directory where archives and checksums are written.")
	return cmd
}
