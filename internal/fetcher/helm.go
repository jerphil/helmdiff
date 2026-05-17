package fetcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type helmFetcher struct {
	repoURL string
	oci     bool
}

func (h *helmFetcher) Pull(chart, version string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "helmdiff-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	args := []string{"pull", chart, "--version", version, "--untar", "--untardir", dir}
	if !h.oci && h.repoURL != "" {
		args = append(args, "--repo", h.repoURL)
	}

	cmd := exec.Command("helm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("helm pull failed: %s\n%s", err, strings.TrimSpace(string(out)))
	}

	chartDir, err := findChartDir(dir)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	return chartDir, cleanup, nil
}

// findChartDir returns the path to the actual chart directory inside the untardir.
// helm --untar creates a single subdirectory named after the chart.
func findChartDir(untardir string) (string, error) {
	entries, err := os.ReadDir(untardir)
	if err != nil {
		return "", fmt.Errorf("reading untar dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			return filepath.Join(untardir, e.Name()), nil
		}
	}
	// Chart.yaml might be directly in untardir (some charts don't wrap)
	if _, err := os.Stat(filepath.Join(untardir, "Chart.yaml")); err == nil {
		return untardir, nil
	}
	return "", fmt.Errorf("could not find chart directory inside %s", untardir)
}
