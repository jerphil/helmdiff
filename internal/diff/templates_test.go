package diff

import (
	"strings"
	"testing"

	"github.com/jerphil/helmdiff/internal/chart"
)

func TestDiffTemplates_NoChanges(t *testing.T) {
	tpl := chart.Template{Name: "deployment.yaml", Content: deploymentYAML("v1.0")}
	results := DiffTemplates([]chart.Template{tpl}, []chart.Template{tpl})
	for _, r := range results {
		if len(r.Changes) > 0 {
			t.Errorf("expected no changes for identical templates, got %+v", r.Changes)
		}
	}
}

func TestDiffTemplates_NewTemplate(t *testing.T) {
	old := []chart.Template{}
	new := []chart.Template{{Name: "hpa.yaml", Content: []byte("apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\n")}}
	results := DiffTemplates(old, new)
	if len(results) != 1 || !results[0].IsNew {
		t.Errorf("expected one new template, got %+v", results)
	}
}

func TestDiffTemplates_RemovedTemplate(t *testing.T) {
	old := []chart.Template{{Name: "pdb.yaml", Content: []byte("kind: PodDisruptionBudget\n")}}
	new := []chart.Template{}
	results := DiffTemplates(old, new)
	if len(results) != 1 || !results[0].IsRemoved {
		t.Errorf("expected one removed template, got %+v", results)
	}
}

func TestDiffTemplates_SemanticChange(t *testing.T) {
	old := []chart.Template{{Name: "deployment.yaml", Content: deploymentYAML("v1.0")}}
	new := []chart.Template{{Name: "deployment.yaml", Content: deploymentYAML("v2.0")}}
	results := DiffTemplates(old, new)
	if len(results) == 0 || len(results[0].Changes) == 0 {
		t.Error("expected changes for modified deployment template")
	}
}

func TestDiffTemplates_TplFileKind(t *testing.T) {
	old := []chart.Template{{Name: "_helpers.tpl", Content: []byte("{{- define \"foo\" -}}bar{{- end -}}")}}
	new := []chart.Template{{Name: "_helpers.tpl", Content: []byte("{{- define \"foo\" -}}baz{{- end -}}")}}
	results := DiffTemplates(old, new)
	if len(results) > 0 && results[0].ResourceKind != "Helper" {
		t.Errorf("_helpers.tpl should have kind Helper, got %q", results[0].ResourceKind)
	}
}

func TestDiffTemplates_UnderscorePrefixIsHelper(t *testing.T) {
	old := []chart.Template{{Name: "_params.tpl", Content: []byte("old content")}}
	new := []chart.Template{{Name: "_params.tpl", Content: []byte("new content")}}
	results := DiffTemplates(old, new)
	if len(results) > 0 && results[0].ResourceKind != "Helper" {
		t.Errorf("_params.tpl should have kind Helper, got %q", results[0].ResourceKind)
	}
}

func TestStripHelmDirectives(t *testing.T) {
	input := []byte(`apiVersion: v1
kind: {{ .Values.kind }}
metadata:
  name: {{ include "chart.name" . }}
`)
	out := stripHelmDirectives(input)
	if strings.Contains(string(out), "{{") {
		t.Errorf("stripHelmDirectives should remove all {{ }} blocks, got:\n%s", out)
	}
}

func TestExtractKind(t *testing.T) {
	content := []byte("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myapp\n")
	kind := extractKind(content)
	if kind != "Deployment" {
		t.Errorf("expected Deployment, got %q", kind)
	}
}

func TestExtractKind_WithHelmDirectives(t *testing.T) {
	// Simple inline directive — kind is parseable after stripping
	content := []byte("apiVersion: apps/v1\nkind: DaemonSet\nmetadata:\n  name: {{ .Release.Name }}\n")
	kind := extractKind(content)
	if kind != "DaemonSet" {
		t.Errorf("expected DaemonSet, got %q", kind)
	}
}

func TestExtractKind_UnknownWhenNoKind(t *testing.T) {
	content := []byte("{{- define \"helper\" -}}some text{{- end -}}")
	kind := extractKind(content)
	if kind != "Unknown" {
		t.Errorf("expected Unknown for template without kind, got %q", kind)
	}
}

func TestExtractName(t *testing.T) {
	content := []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: my-service\n")
	name := extractName(content)
	if name != "my-service" {
		t.Errorf("expected my-service, got %q", name)
	}
}

func TestExtractName_WithHelmDirectives(t *testing.T) {
	content := []byte("kind: Service\nmetadata:\n  name: {{ include \"chart.fullname\" . }}\n")
	// name is a template expression — should return empty (not "__helm__")
	name := extractName(content)
	if name == "__helm__" {
		t.Errorf("extractName should not return __helm__ placeholder, got %q", name)
	}
}

func TestIsTplHelper(t *testing.T) {
	cases := []struct {
		name     string
		expected bool
	}{
		{"_helpers.tpl", true},
		{"_params.tpl", true},
		{"deployment.yaml", false},
		{"templates/_helpers.tpl", true},
		{"templates/deployment.yaml", false},
		{"something.tpl", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isTplHelper(tc.name)
			if got != tc.expected {
				t.Errorf("isTplHelper(%q) = %v, want %v", tc.name, got, tc.expected)
			}
		})
	}
}

func TestSplitYAMLDocuments(t *testing.T) {
	input := []byte("doc1: true\n---\ndoc2: true")
	docs := splitYAMLDocuments(input)
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}
}

func TestSplitYAMLDocuments_SingleDoc(t *testing.T) {
	input := []byte("key: value")
	docs := splitYAMLDocuments(input)
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}
}

// deploymentYAML returns a minimal Deployment YAML with the given image tag.
func deploymentYAML(tag string) []byte {
	return []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:` + tag + `
`)
}
