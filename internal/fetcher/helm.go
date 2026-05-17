package fetcher

import (
	"fmt"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
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

	rc, err := registry.NewClient()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("creating registry client: %w", err)
	}

	cfg := new(action.Configuration)
	cfg.RegistryClient = rc

	pullClient := action.NewPullWithOpts(action.WithConfig(cfg))
	pullClient.Settings = cli.New()
	pullClient.Version = version
	pullClient.Untar = true
	pullClient.UntarDir = dir

	if !h.oci && h.repoURL != "" {
		pullClient.RepoURL = h.repoURL
	}

	if _, err := pullClient.Run(chart); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("pulling chart: %w", err)
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
