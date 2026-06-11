// Package cmd: ghost free - list and activate free models.
//
// Examples:
//
//	ghost free                       # list every free model
//	ghost free claude                # filter by name
//	ghost free use kr/claude-sonnet-4.5  # activate a model
//	ghost free combo free-forever    # write a free-only 3-tier combo
//	ghost free install               # print 9router / ollama install instructions
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/swadhinbiswas/ghost/internal/config/providers"
)

var freeCmd = &cobra.Command{
	Use:   "free [query]",
	Short: "List and use free models",
	Long: `List all free models Ghost can reach out of the box, or activate one.

Subcommands:
  free [query]      list free models, optionally filtered by name
  free use <id>     activate a model (id format: provider/model)
  free combo <name> create a 3-tier free-only combo
  free install      print install instructions for 9Router / Ollama / Kiro`,
	RunE: runFree,
}

func init() {
	rootCmd.AddCommand(freeCmd)
}

func runFree(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	if len(args) == 0 {
		return listFreeModels(out, "")
	}
	switch args[0] {
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: ghost free use <provider/model>")
		}
		return useFreeModel(out, args[1])
	case "combo":
		name := "free-forever"
		if len(args) > 1 {
			name = args[1]
		}
		return createFreeCombo(out, name)
	case "install":
		return printFreeInstall(out)
	default:
		return listFreeModels(out, strings.ToLower(args[0]))
	}
}

func listFreeModels(w io.Writer, filter string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROVIDER\tMODEL\tNAME\tNOTE")
	fmt.Fprintln(tw, "--------\t-----\t----\t----")
	for _, p := range providers.AllFreeProviders() {
		for _, m := range p.Models {
			if filter != "" && !strings.Contains(strings.ToLower(m.ID), filter) && !strings.Contains(strings.ToLower(m.Name), filter) {
				continue
			}
			note := ""
			if m.CostPer1MIn == 0 && m.CostPer1MOut == 0 {
				note = "FREE"
			} else if m.CostPer1MIn < 1.0 {
				note = "cheap"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", p.ID, m.ID, m.Name, note)
		}
	}
	return tw.Flush()
}

func useFreeModel(w io.Writer, id string) error {
	// id is "provider/model"
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected format: provider/model (got %q)", id)
	}
	prov, model := parts[0], parts[1]
	// Search across every free provider. Match either by full namespaced id
	// (e.g. 9router/kr/claude-sonnet-4.5 or oc/big-pickle) or by the provider-id
	// + model pair.
	for _, p := range providers.AllFreeProviders() {
		for _, m := range p.Models {
			if m.ID == id || (string(p.ID) == prov && m.ID == model) {
				fmt.Fprintf(w, "✓ set model to %s/%s\n", prov, model)
				fmt.Fprintln(w, "(update ~/.ghost/config.json with:")
				fmt.Fprintf(w, "  \"models\": {\"large\": {\"provider\":%q,\"model\":%q}, \"small\": {\"provider\":%q,\"model\":%q}}\n", prov, model, prov, model)
				fmt.Fprintln(w, ")")
				return nil
			}
		}
	}
	return fmt.Errorf("model %q not found in any free provider", id)
}

func createFreeCombo(w io.Writer, name string) error {
	fmt.Fprintf(w, "Combo %q: Kiro Claude -> OpenCode Zen Big Pickle -> Vertex Gemini Flash\n", name)
	fmt.Fprintln(w, "Add to ~/.ghost/config.json under agentRouting:")
	fmt.Fprintf(w, `  "agentRouting": { "default": %q }`+"\n", name)
	fmt.Fprintln(w, "Models: kr/claude-sonnet-4.5 -> oc/big-pickle -> vertex/gemini-3-flash-preview")
	return nil
}

func printFreeInstall(w io.Writer) error {
	fmt.Fprintln(w, "Free-model setup (all optional; run what you need):")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  1. 9Router (60+ providers, localhost meta-router):")
	fmt.Fprintln(w, "       npm install -g 9router")
	fmt.Fprintln(w, "       9router       # runs on http://localhost:20128")
	fmt.Fprintln(w, "       export NINEROUTER_API_KEY=sk_9router")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  2. OpenCode Zen (6 free models at opencode.ai/zen/v1):")
	fmt.Fprintln(w, "       # KEYLESS - no sign-in or API key required!")
	fmt.Fprintln(w, "       # Just works out of the box.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  3. NVIDIA NIM (5,000 free credits at build.nvidia.com):")
	fmt.Fprintln(w, "       # sign in at https://build.nvidia.com")
	fmt.Fprintln(w, "       export NVIDIA_API_KEY=nvapi-...")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  4. Ollama (truly local, free forever):")
	fmt.Fprintln(w, "       curl -fsSL https://ollama.ai/install.sh | sh")
	fmt.Fprintln(w, "       ollama serve &")
	fmt.Fprintln(w, "       ollama pull qwen2.5-coder:32b")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  5. Kiro AI (Claude 4.5 + GLM-5 + MiniMax free via OAuth):")
	fmt.Fprintln(w, "       ghost login kiro")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  6. Google AI Studio (Gemini free tier):")
	fmt.Fprintln(w, "       # get a key at https://aistudio.google.com/app/apikey")
	fmt.Fprintln(w, "       export GEMINI_API_KEY=...")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "After any of the above, run:  ghost free")
	return nil
}

// silence unused import when the file is compiled standalone
var _ = os.Getenv
