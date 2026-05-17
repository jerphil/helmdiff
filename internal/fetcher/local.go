package fetcher

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type localFetcher struct{}

// Pull for a local fetcher ignores the version and treats chart as a file/dir path.
func (l *localFetcher) Pull(path, _ string) (string, func(), error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if info.IsDir() {
		return path, func() {}, nil
	}

	if !strings.HasSuffix(path, ".tgz") && !strings.HasSuffix(path, ".tar.gz") {
		return "", nil, fmt.Errorf("%s is not a directory or .tgz file", path)
	}

	dir, err := os.MkdirTemp("", "helmdiff-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	if err := extractTGZ(path, dir); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("extracting %s: %w", path, err)
	}

	chartDir, err := findChartDir(dir)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	return chartDir, cleanup, nil
}

func extractTGZ(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, hdr.Name)
		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec
				_ = out.Close()
				return err
			}
			_ = out.Close()
		}
	}
	return nil
}
