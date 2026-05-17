package fetcher

import "strings"

type Fetcher interface {
	Pull(chart, version string) (chartDir string, cleanup func(), err error)
}

func New(chart, repoURL string) Fetcher {
	if strings.HasPrefix(chart, "oci://") {
		return &helmFetcher{repoURL: "", oci: true}
	}
	if repoURL == "" {
		repoURL = resolve(chart)
	}
	return &helmFetcher{repoURL: repoURL}
}
