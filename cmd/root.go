package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jerphil/helmdiff/internal/ai"
	"github.com/jerphil/helmdiff/internal/chart"
	"github.com/jerphil/helmdiff/internal/diff"
	"github.com/jerphil/helmdiff/internal/fetcher"
	"github.com/jerphil/helmdiff/internal/renderer"
	"github.com/spf13/cobra"
)

var (
	flagRepo    string
	flagOutput  string
	flagAI      bool
	flagAIModel string
	flagFailOn  string
)

// SetVersion is called by main with values injected via -ldflags at build time.
func SetVersion(v, c, d string) {
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}

var rootCmd = &cobra.Command{
	Use:   "helmdiff [chart] [old-version] [new-version]",
	Short: "Diff two versions of a Helm chart before touching a cluster",
	Long: `helmdiff pulls two versions of a Helm chart from any registry and
produces a human-readable diff of templates, values, CRDs, and chart metadata.
Each change is classified with a risk level: CRITICAL, HIGH, MEDIUM, or LOW.

Well-known charts (ingress-nginx, cert-manager, kube-prometheus-stack, argo-cd,
and 40+ others) are resolved automatically — no --repo flag needed.
No external dependencies required — helm does not need to be installed.

Examples:
  helmdiff ingress-nginx 4.9.0 4.11.0
  helmdiff cert-manager 1.13.0 1.15.0
  helmdiff my-chart 1.0.0 2.0.0 --repo https://my-org.github.io/charts
  helmdiff oci://registry.k8s.io/ingress-nginx/ingress-nginx 4.9.0 4.11.0
  helmdiff ingress-nginx 4.9.0 4.11.0 -o json | jq '.resources[].changes[]'
  helmdiff cert-manager 1.13.0 1.15.0 --ai
  helmdiff ingress-nginx 4.9.0 4.11.0 --fail-on high   # exit 1 if HIGH or CRITICAL changes found
  helmdiff ./chart-v1.tgz ./chart-v2.tgz              # diff two local .tgz files
  helmdiff ./old-chart/ ./new-chart/                   # diff two local directories

Environment variables:
  HELMDIFF_AI_API_KEY    API key for the AI provider (required with --ai)
  HELMDIFF_AI_BASE_URL   Base URL of any OpenAI-compatible endpoint
                           Anthropic (default): https://api.anthropic.com/v1
                           OpenAI:              https://api.openai.com/v1
                           OpenRouter:          https://openrouter.ai/api/v1
                           Ollama (local):      http://localhost:11434/v1
  HELMDIFF_AI_MODEL      Model to use (default: claude-sonnet-4-6)
                           Overridden by --ai-model if both are set`,
	Args: cobra.RangeArgs(2, 3),
	RunE: run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVar(&flagRepo, "repo", "", "Helm repository URL (auto-detected for known charts)")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "human", "Output format: human, json")
	rootCmd.Flags().BoolVar(&flagAI, "ai", false, "Summarize breaking changes with AI (requires HELMDIFF_AI_API_KEY)")
	rootCmd.Flags().StringVar(&flagAIModel, "ai-model", "", "Model override (e.g. gpt-4o, claude-sonnet-4-6, llama3)")
	rootCmd.Flags().StringVar(&flagFailOn, "fail-on", "", "Exit with code 1 if changes at or above this risk level are found (low, medium, high, critical)")
}

func run(cmd *cobra.Command, args []string) error {
	// Local mode: helmdiff ./old.tgz ./new.tgz
	if len(args) == 2 {
		return runLocal(args[0], args[1])
	}

	chartName := args[0]
	oldVersion := args[1]
	newVersion := args[2]

	if oldVersion == newVersion {
		return fmt.Errorf("old and new versions are identical: %s", oldVersion)
	}

	f := fetcher.New(chartName, flagRepo)

	fmt.Fprintf(os.Stderr, "Pulling %s %s...\n", chartName, oldVersion)
	oldDir, cleanupOld, err := f.Pull(chartName, oldVersion)
	if err != nil {
		return fmt.Errorf("fetching old version: %w", err)
	}
	defer cleanupOld()

	fmt.Fprintf(os.Stderr, "Pulling %s %s...\n", chartName, newVersion)
	newDir, cleanupNew, err := f.Pull(chartName, newVersion)
	if err != nil {
		return fmt.Errorf("fetching new version: %w", err)
	}
	defer cleanupNew()

	fmt.Fprintln(os.Stderr, "Comparing charts...")
	return diffAndRender(oldDir, newDir, chartName)
}

func diffAndRender(oldDir, newDir, chartName string) error {
	oldChart, err := chart.Load(oldDir)
	if err != nil {
		return fmt.Errorf("loading old chart: %w", err)
	}
	newChart, err := chart.Load(newDir)
	if err != nil {
		return fmt.Errorf("loading new chart: %w", err)
	}

	report := diff.Run(oldChart, newChart)
	if report.ChartName == "" {
		report.ChartName = chartName
	}

	var r renderer.Renderer
	switch strings.ToLower(flagOutput) {
	case "json":
		r = &renderer.JSONRenderer{}
	case "human", "":
		r = &renderer.HumanRenderer{}
	default:
		return fmt.Errorf("unknown output format %q (use: human, json)", flagOutput)
	}

	if err := r.Render(report); err != nil {
		return err
	}

	if flagAI {
		if err := ai.Summarize(report, flagAIModel); err != nil {
			fmt.Fprintf(os.Stderr, "AI summary failed: %v\n", err)
		}
	}

	if flagFailOn != "" {
		threshold, err := parseRiskLevel(flagFailOn)
		if err != nil {
			return err
		}
		if report.MaxRisk() >= threshold {
			return fmt.Errorf("changes at or above %s risk level found", strings.ToUpper(flagFailOn))
		}
	}

	return nil
}

func runLocal(oldPath, newPath string) error {
	lf := fetcher.New(oldPath, "")

	fmt.Fprintf(os.Stderr, "Loading %s...\n", oldPath)
	oldDir, cleanupOld, err := lf.Pull(oldPath, "")
	if err != nil {
		return fmt.Errorf("loading old chart: %w", err)
	}
	defer cleanupOld()

	fmt.Fprintf(os.Stderr, "Loading %s...\n", newPath)
	newDir, cleanupNew, err := fetcher.New(newPath, "").Pull(newPath, "")
	if err != nil {
		return fmt.Errorf("loading new chart: %w", err)
	}
	defer cleanupNew()

	return diffAndRender(oldDir, newDir, oldPath)
}

func parseRiskLevel(s string) (diff.RiskLevel, error) {
	switch strings.ToLower(s) {
	case "low":
		return diff.RiskLow, nil
	case "medium":
		return diff.RiskMedium, nil
	case "high":
		return diff.RiskHigh, nil
	case "critical":
		return diff.RiskCritical, nil
	default:
		return diff.RiskLow, fmt.Errorf("unknown risk level %q (use: low, medium, high, critical)", s)
	}
}
