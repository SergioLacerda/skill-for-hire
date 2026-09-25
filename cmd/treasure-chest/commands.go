package main

import (
	"encoding/json"
	"fmt"
	"strings"

	ka "github.com/SergioLacerda/skill-for-hire/platform/knowledge-api"
	"github.com/SergioLacerda/skill-for-hire/skills/treasure-chest/runtime"
	"github.com/spf13/cobra"
)

// asJSON gates every command's --json output. Uses json.MarshalIndent
// so the human eye can also read it.
type outputMode struct {
	json bool
}

func (o *outputMode) attach(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.json, "json", false, "Emit machine-readable JSON instead of human text.")
}

func (o *outputMode) render(cmd *cobra.Command, payload any) error {
	if o.json {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	}
	// Human render: fmt.Fprintln so callers still see structured output.
	_, err := fmt.Fprintln(cmd.OutOrStdout(), payload)
	return err
}

func newPrepareCmd(a *runtime.Adapter) *cobra.Command {
	var root string
	out := &outputMode{}
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Load governed chests + jewels + potions under --root and build the in-memory index.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := a.Prepare(cmd.Context(), ka.PrepareRequest{Root: root})
			if err != nil {
				return err
			}
			if out.json {
				return out.render(cmd, res)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(),
				"prepared root=%s sources=%d items=%d provider=%s@%s\n",
				root, res.SourcesRead, res.ItemsIndexed,
				res.Envelope.Provider, res.Envelope.Version)
			return err
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "Workspace root to load treasure-chest artifacts from.")
	out.attach(cmd)
	return cmd
}

func newSearchCmd(a *runtime.Adapter) *cobra.Command {
	var (
		root        string
		intent      string
		concepts    []string
		symbols     []string
		tokenBudget int
	)
	out := &outputMode{}
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search prepared jewels/potions by intent, concepts, or symbols.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Auto-prepare so a bare `treasure-chest search` works
			// out of the box against the given root.
			if _, err := a.Prepare(cmd.Context(), ka.PrepareRequest{Root: root}); err != nil {
				return err
			}
			res, err := a.Search(cmd.Context(), ka.Query{
				Intent:      intent,
				Concepts:    concepts,
				Symbols:     symbols,
				TokenBudget: tokenBudget,
			})
			if err != nil {
				return err
			}
			if out.json {
				return out.render(cmd, res)
			}
			w := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(w, "%d item(s) matched\n", len(res.Items)); err != nil {
				return err
			}
			for _, it := range res.Items {
				if _, err := fmt.Fprintf(w, "  %-8s %-40s applicability=%s reason=%s\n",
					it.Kind, it.ID, it.Applicability, it.Reason); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "Workspace root.")
	cmd.Flags().StringVar(&intent, "intent", "", "High-level intent to match against jewel statements.")
	cmd.Flags().StringSliceVar(&concepts, "concept", nil, "Concept keyword to match; repeatable.")
	cmd.Flags().StringSliceVar(&symbols, "symbol", nil, "Symbol keyword to match; repeatable.")
	cmd.Flags().IntVar(&tokenBudget, "token-budget", 0, "Cap the number of returned items (0 = unlimited).")
	out.attach(cmd)
	return cmd
}

func newStatusCmd(a *runtime.Adapter) *cobra.Command {
	var root string
	out := &outputMode{}
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report provider identity, health, capabilities and index counts.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Best-effort: attempt Prepare so counts are populated. A
			// Prepare failure does not fail Status — the point of
			// Status is to reveal that.
			_, _ = a.Prepare(cmd.Context(), ka.PrepareRequest{Root: root})
			s, err := a.Status(cmd.Context())
			if err != nil {
				return err
			}
			if out.json {
				return out.render(cmd, s)
			}
			w := cmd.OutOrStdout()
			healthy := "no"
			if s.Healthy {
				healthy = "yes"
			}
			_, err = fmt.Fprintf(w,
				"provider:      %s@%s\nschema:        %d\nhealthy:       %s\ncapabilities:  %s\ndetail:        %s\n",
				s.Provider, s.Version, s.SchemaVersion, healthy,
				strings.Join(s.Capabilities, ", "), s.Detail)
			return err
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "Workspace root.")
	out.attach(cmd)
	return cmd
}

func newExplainCmd(a *runtime.Adapter) *cobra.Command {
	var root string
	out := &outputMode{}
	cmd := &cobra.Command{
		Use:   "explain <item-id>",
		Short: "Explain why an item was selected/discarded from the current index.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := a.Prepare(cmd.Context(), ka.PrepareRequest{Root: root}); err != nil {
				return err
			}
			exp, err := a.Explain(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if out.json {
				return out.render(cmd, exp)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(),
				"item:     %s\ndecision: %s\nreason:   %s\nevidence: %d source(s)\n",
				exp.ItemID, exp.Decision, exp.Reason, len(exp.Evidence))
			return err
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "Workspace root.")
	out.attach(cmd)
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the treasure-chest CLI version.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), Version)
			return err
		},
	}
}
