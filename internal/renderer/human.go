package renderer

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/jerphil/helmdiff/internal/diff"
)

var (
	critical = color.New(color.FgHiRed, color.Bold)
	high     = color.New(color.FgRed)
	medium   = color.New(color.FgYellow)
	low      = color.New(color.FgWhite)
	header   = color.New(color.FgCyan, color.Bold)
	bold     = color.New(color.Bold)
	dim      = color.New(color.FgHiBlack)
	green    = color.New(color.FgGreen)
	added    = color.New(color.FgGreen)
	removed  = color.New(color.FgRed)
)

type HumanRenderer struct{}

func (h *HumanRenderer) Render(report *diff.DiffReport) error {
	printHeader(report)
	printSummary(report)
	printMetaChanges(report)
	printCRDChanges(report)
	printResourceChanges(report)
	printValueChanges(report)
	return nil
}

func printHeader(r *diff.DiffReport) {
	fmt.Fprintln(os.Stdout)
	header.Fprintf(os.Stdout, "  helmdiff: %s\n", r.ChartName)
	dim.Fprintf(os.Stdout, "  %s → %s  (generated %s)\n\n", r.OldVersion, r.NewVersion, r.GeneratedAt.Format("2006-01-02 15:04:05"))
}

func printSummary(r *diff.DiffReport) {
	highCount := r.HighCount()
	medCount := r.MediumCount()
	lowCount := r.LowCount()
	total := highCount + medCount + lowCount

	if total == 0 {
		green.Fprintln(os.Stdout, "  No changes detected.")
		fmt.Fprintln(os.Stdout)
		return
	}

	bold.Fprintln(os.Stdout, "  Summary")
	fmt.Fprintln(os.Stdout, "  "+strings.Repeat("─", 50))

	if len(r.CRDChanges) > 0 {
		critical.Fprintf(os.Stdout, "  %-12s %d CRD change(s)\n", "[CRITICAL]", len(r.CRDChanges))
	}
	if highCount > 0 {
		high.Fprintf(os.Stdout, "  %-12s %d change(s)\n", "[HIGH]", highCount)
	}
	if medCount > 0 {
		medium.Fprintf(os.Stdout, "  %-12s %d change(s)\n", "[MEDIUM]", medCount)
	}
	if lowCount > 0 {
		low.Fprintf(os.Stdout, "  %-12s %d change(s)\n", "[LOW]", lowCount)
	}
	fmt.Fprintln(os.Stdout, "  "+strings.Repeat("─", 50))
	fmt.Fprintf(os.Stdout, "  Total: %d change(s) across %d template file(s)\n\n", total, len(r.Resources))
}

func printMetaChanges(r *diff.DiffReport) {
	if len(r.MetaChanges) == 0 {
		return
	}
	printSection("Chart.yaml", "", r.MetaChanges)
}

func printCRDChanges(r *diff.DiffReport) {
	if len(r.CRDChanges) == 0 {
		return
	}
	printSection("CRDs", "", r.CRDChanges)
}

func printResourceChanges(r *diff.DiffReport) {
	// Group by kind
	byKind := make(map[string][]diff.ResourceDiff)
	for _, res := range r.Resources {
		kind := res.ResourceKind
		if kind == "" {
			kind = "Unknown"
		}
		byKind[kind] = append(byKind[kind], res)
	}

	// Sort kinds for deterministic output
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)

	for _, kind := range kinds {
		resources := byKind[kind]
		for _, res := range resources {
			if res.IsNew {
				header.Fprintf(os.Stdout, "  %s/%s", kind, res.TemplateFile)
				added.Fprintln(os.Stdout, "  [NEW TEMPLATE]")
				fmt.Fprintln(os.Stdout)
				continue
			}
			if res.IsRemoved {
				header.Fprintf(os.Stdout, "  %s/%s", kind, res.TemplateFile)
				removed.Fprintln(os.Stdout, "  [TEMPLATE REMOVED]")
				fmt.Fprintln(os.Stdout)
				continue
			}
			if len(res.Changes) == 0 {
				continue
			}
			name := res.ResourceName
			if name == "" {
				name = res.TemplateFile
			}
			printSection(kind, name, res.Changes)
		}
	}
}

func printValueChanges(r *diff.DiffReport) {
	if len(r.ValueChanges) == 0 {
		return
	}
	printSection("values.yaml", "", r.ValueChanges)
}

func printSection(kind, name string, changes []diff.Change) {
	title := kind
	if name != "" {
		title = kind + " / " + name
	}
	header.Fprintf(os.Stdout, "  %s\n", title)
	fmt.Fprintln(os.Stdout, "  "+strings.Repeat("─", 50))

	for _, c := range changes {
		printChange(c)
	}
	fmt.Fprintln(os.Stdout)
}

func printChange(c diff.Change) {
	risk := riskColor(c.Risk)
	badge := fmt.Sprintf("[%-8s]", c.Risk.String())

	if c.Path == "(raw diff)" {
		risk.Fprintf(os.Stdout, "  %s ", badge)
		medium.Fprintln(os.Stdout, c.Description)
		if diff, ok := c.NewValue.(string); ok {
			printRawDiff(diff)
		}
		return
	}

	risk.Fprintf(os.Stdout, "  %s ", badge)
	switch c.Kind {
	case diff.Added:
		added.Fprintf(os.Stdout, "+ %s: ", c.Path)
		fmt.Fprintf(os.Stdout, "%v\n", formatValue(c.NewValue))
	case diff.Removed:
		removed.Fprintf(os.Stdout, "- %s: ", c.Path)
		fmt.Fprintf(os.Stdout, "%v\n", formatValue(c.OldValue))
	case diff.Changed:
		bold.Fprintf(os.Stdout, "~ %s: ", c.Path)
		removed.Fprintf(os.Stdout, "%v", formatValue(c.OldValue))
		fmt.Fprintf(os.Stdout, " → ")
		added.Fprintf(os.Stdout, "%v\n", formatValue(c.NewValue))
	}
}

func printRawDiff(d string) {
	for _, line := range strings.Split(d, "\n") {
		if strings.HasPrefix(line, "+") {
			added.Fprintf(os.Stdout, "    %s\n", line)
		} else if strings.HasPrefix(line, "-") {
			removed.Fprintf(os.Stdout, "    %s\n", line)
		} else {
			dim.Fprintf(os.Stdout, "    %s\n", line)
		}
	}
}

func riskColor(r diff.RiskLevel) *color.Color {
	switch r {
	case diff.RiskCritical:
		return critical
	case diff.RiskHigh:
		return high
	case diff.RiskMedium:
		return medium
	default:
		return low
	}
}

func formatValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	s := fmt.Sprintf("%v", v)
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}
