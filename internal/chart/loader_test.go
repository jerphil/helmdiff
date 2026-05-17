package chart

import (
	"os"
	"path/filepath"
	"testing"
)

func makeChart(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestLoad_Basic(t *testing.T) {
	dir := makeChart(t, map[string]string{
		"Chart.yaml":  "name: mychart\nversion: 1.0.0\nappVersion: \"1.2.3\"\n",
		"values.yaml": "replicaCount: 2\n",
		"templates/deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\n",
	})

	c, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if c.Meta.Name != "mychart" {
		t.Errorf("expected name mychart, got %s", c.Meta.Name)
	}
	if c.Meta.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", c.Meta.Version)
	}
	if c.Meta.AppVersion != "1.2.3" {
		t.Errorf("expected appVersion 1.2.3, got %s", c.Meta.AppVersion)
	}
	if len(c.Templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(c.Templates))
	}
	if v, ok := c.Values["replicaCount"]; !ok || v != 2 {
		t.Errorf("expected replicaCount=2, got %v", v)
	}
}

func TestLoad_NoValues(t *testing.T) {
	dir := makeChart(t, map[string]string{
		"Chart.yaml": "name: mychart\nversion: 1.0.0\n",
	})
	c, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if c.Values != nil {
		t.Errorf("expected nil values, got %v", c.Values)
	}
}

func TestLoad_SubdirWrapping(t *testing.T) {
	outer := t.TempDir()
	inner := filepath.Join(outer, "mychart")
	if err := os.Mkdir(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "Chart.yaml"), []byte("name: mychart\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := Load(outer)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if c.Meta.Name != "mychart" {
		t.Errorf("expected name mychart, got %s", c.Meta.Name)
	}
}

func TestLoad_WithCRD(t *testing.T) {
	dir := makeChart(t, map[string]string{
		"Chart.yaml": "name: mychart\nversion: 1.0.0\n",
		"crds/myresource.yaml": `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myresources.example.com
spec:
  group: example.com
  names:
    kind: MyResource
  versions:
    - name: v1
`,
	})

	c, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(c.CRDs) != 1 {
		t.Fatalf("expected 1 CRD, got %d", len(c.CRDs))
	}
	if c.CRDs[0].Kind != "MyResource" {
		t.Errorf("expected kind MyResource, got %s", c.CRDs[0].Kind)
	}
	if c.CRDs[0].Group != "example.com" {
		t.Errorf("expected group example.com, got %s", c.CRDs[0].Group)
	}
}

func TestLoad_MissingChartYaml(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for missing Chart.yaml")
	}
}

func TestLoad_TemplateExtensions(t *testing.T) {
	dir := makeChart(t, map[string]string{
		"Chart.yaml":              "name: mychart\nversion: 1.0.0\n",
		"templates/deploy.yaml":   "kind: Deployment",
		"templates/_helpers.tpl":  "{{- define \"helper\" -}}{{- end -}}",
		"templates/notes.txt":     "this should be ignored",
	})

	c, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(c.Templates) != 2 {
		t.Errorf("expected 2 templates (yaml+tpl, not txt), got %d", len(c.Templates))
	}
}
