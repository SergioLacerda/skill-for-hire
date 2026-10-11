package main

import (
	"fmt"
	"io"

	"github.com/SergioLacerda/skill-for-hire/internal/skill"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var rosterPath string
	cmd := &cobra.Command{
		Use:   "validate <skill-dir> [<skill-dir>...]",
		Short: "Package-static validation of a skill directory.",
		Long: `Package-static validation of one or more skill directories.

With --roster, every skill is also cross-checked against the catalog file:
it must be listed there and its roster version must equal the manifest's
metadata.version.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			var failed int
			for _, dir := range args {
				report, err := skill.Validate(dir)
				if err != nil {
					if _, werr := fmt.Fprintf(out, "%s: %v\n", dir, err); werr != nil {
						return fmt.Errorf("write report: %w", werr)
					}
					failed++
					continue
				}
				if err := printReport(out, report); err != nil {
					return err
				}
				if !report.OK() {
					failed++
				}
			}
			if failed > 0 {
				return fmt.Errorf("%d skill(s) failed package-static validation", failed)
			}
			if rosterPath == "" {
				return nil
			}
			return checkRoster(out, rosterPath, args)
		},
	}
	cmd.Flags().StringVar(&rosterPath, "roster", "",
		"catalog file (roster/skills.yaml) to cross-check skill versions against")
	return cmd
}

// checkRoster prints every roster disagreement and fails when any exists.
func checkRoster(out io.Writer, rosterPath string, skillDirs []string) error {
	problems, err := skill.CheckRoster(rosterPath, skillDirs)
	if err != nil {
		return fmt.Errorf("roster cross-check: %w", err)
	}
	for _, p := range problems {
		if _, werr := fmt.Fprintf(out, "FAIL roster: %s\n", p); werr != nil {
			return fmt.Errorf("write roster report: %w", werr)
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d roster problem(s): catalog and manifests disagree", len(problems))
	}
	return nil
}

func printReport(out io.Writer, r *skill.Report) error {
	name := r.SkillDir
	if r.Manifest != nil && r.Manifest.Metadata.Name != "" {
		name = fmt.Sprintf("%s (%s@%s)", r.SkillDir, r.Manifest.Metadata.Name, r.Manifest.Metadata.Version)
	}
	status := "ok  "
	if !r.OK() {
		status = "FAIL"
	}
	if _, err := fmt.Fprintf(out, "%s %s\n", status, name); err != nil {
		return fmt.Errorf("write report header: %w", err)
	}
	for _, w := range r.Warnings {
		if _, err := fmt.Fprintf(out, "  warn: %s\n", w); err != nil {
			return fmt.Errorf("write warning: %w", err)
		}
	}
	for _, e := range r.Errors {
		if _, err := fmt.Fprintf(out, "  err : %s\n", e); err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}
	return nil
}
