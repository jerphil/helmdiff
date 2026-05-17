package fetcher

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestIsLocalPath(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"./chart", true},
		{"../chart", true},
		{"/abs/path", true},
		{"chart.tgz", true},
		{"chart.tar.gz", true},
		{"ingress-nginx", false},
		{"oci://registry.k8s.io/foo", false},
	}
	for _, tc := range cases {
		if got := isLocalPath(tc.input); got != tc.want {
			t.Errorf("isLocalPath(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsLocalPath_ExistingDir(t *testing.T) {
	dir := t.TempDir()
	if !isLocalPath(dir) {
		t.Errorf("isLocalPath(%q) should be true for existing directory", dir)
	}
}

func TestLocalFetcher_Directory(t *testing.T) {
	dir := t.TempDir()
	lf := &localFetcher{}
	got, cleanup, err := lf.Pull(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if got != dir {
		t.Errorf("expected %q, got %q", dir, got)
	}
}

func TestLocalFetcher_TGZ(t *testing.T) {
	tgz := makeTGZ(t, "mychart", map[string]string{
		"Chart.yaml":  "apiVersion: v2\nname: mychart\nversion: 1.0.0\n",
		"values.yaml": "key: val\n",
	})

	lf := &localFetcher{}
	chartDir, cleanup, err := lf.Pull(tgz, "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if _, err := os.Stat(filepath.Join(chartDir, "Chart.yaml")); err != nil {
		t.Errorf("Chart.yaml not found in extracted dir: %v", err)
	}
}

func TestLocalFetcher_InvalidPath(t *testing.T) {
	lf := &localFetcher{}
	_, _, err := lf.Pull("/nonexistent/path.tgz", "")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

// makeTGZ creates a temporary .tgz with files under a chart subdirectory.
func makeTGZ(t *testing.T, chartName string, files map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.tgz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		body := []byte(content)
		hdr := &tar.Header{
			Name: chartName + "/" + name,
			Mode: 0o644,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}
