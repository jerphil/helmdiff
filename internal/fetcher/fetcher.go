package fetcher

import (
	"os"
	"strings"
)

// Fetcher pulls a chart version and returns its unpacked directory path.
type Fetcher interface {
	Pull(chart, version string) (chartDir string, cleanup func(), err error)
}

// New returns a Fetcher appropriate for the given chart reference.
// Local paths (.tgz files or directories) are served without network access.
func New(chart, repoURL string) Fetcher {
	if isLocalPath(chart) {
		return &localFetcher{}
	}
	if strings.HasPrefix(chart, "oci://") {
		return &helmFetcher{repoURL: "", oci: true}
	}
	if repoURL == "" {
		repoURL = resolve(chart)
	}
	return &helmFetcher{repoURL: repoURL}
}

func isLocalPath(chart string) bool {
	if strings.HasPrefix(chart, "./") || strings.HasPrefix(chart, "/") || strings.HasPrefix(chart, "../") {
		return true
	}
	if strings.HasSuffix(chart, ".tgz") || strings.HasSuffix(chart, ".tar.gz") {
		return true
	}
	if info, err := os.Stat(chart); err == nil && info.IsDir() {
		return true
	}
	return false
}
