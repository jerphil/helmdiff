package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/jerphil/helmdiff/internal/diff"
)

func makeReport(changes []diff.Change) *diff.DiffReport {
	return &diff.DiffReport{
		ChartName:   "mychart",
		OldVersion:  "1.0.0",
		NewVersion:  "2.0.0",
		GeneratedAt: time.Now(),
		MetaChanges: changes,
	}
}

func TestHumanRenderer_NoChanges(t *testing.T) {
	out := captureStdout(t, func() {
		r := &HumanRenderer{}
		if err := r.Render(makeReport(nil)); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "No changes detected") {
		t.Errorf("expected 'No changes detected', got:\n%s", out)
	}
}

func TestHumanRenderer_WithChanges(t *testing.T) {
	changes := []diff.Change{
		{Path: "appVersion", Kind: diff.Changed, OldValue: "1.0", NewValue: "2.0", Risk: diff.RiskMedium, Description: "appVersion changed"},
		{Path: "image.tag", Kind: diff.Changed, OldValue: "v1", NewValue: "v2", Risk: diff.RiskHigh, Description: "image tag changed"},
	}
	out := captureStdout(t, func() {
		r := &HumanRenderer{}
		if err := r.Render(makeReport(changes)); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "HIGH") {
		t.Errorf("expected HIGH in output, got:\n%s", out)
	}
	if !strings.Contains(out, "MEDIUM") {
		t.Errorf("expected MEDIUM in output, got:\n%s", out)
	}
	if !strings.Contains(out, "appVersion") {
		t.Errorf("expected appVersion in output, got:\n%s", out)
	}
}

func TestHumanRenderer_CRDChanges(t *testing.T) {
	report := &diff.DiffReport{
		ChartName:   "mychart",
		OldVersion:  "1.0.0",
		NewVersion:  "2.0.0",
		GeneratedAt: time.Now(),
		CRDChanges:  []diff.Change{{Path: "myresource", Kind: diff.Removed, Risk: diff.RiskCritical, Description: "CRD removed"}},
	}
	out := captureStdout(t, func() {
		r := &HumanRenderer{}
		if err := r.Render(report); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "CRITICAL") {
		t.Errorf("expected CRITICAL in output, got:\n%s", out)
	}
}

func TestHumanRenderer_AddedRemoved(t *testing.T) {
	changes := []diff.Change{
		{Path: "key", Kind: diff.Added, NewValue: "val", Risk: diff.RiskLow},
		{Path: "old", Kind: diff.Removed, OldValue: "val", Risk: diff.RiskLow},
	}
	out := captureStdout(t, func() {
		r := &HumanRenderer{}
		if err := r.Render(makeReport(changes)); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "+ key") {
		t.Errorf("expected '+ key' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "- old") {
		t.Errorf("expected '- old' in output, got:\n%s", out)
	}
}

func TestHumanRenderer_RawDiff(t *testing.T) {
	changes := []diff.Change{
		{Path: "(raw diff)", Kind: diff.Changed, NewValue: "+added line\n-removed line", Risk: diff.RiskMedium, Description: "template changed"},
	}
	out := captureStdout(t, func() {
		r := &HumanRenderer{}
		if err := r.Render(makeReport(changes)); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "template changed") {
		t.Errorf("expected description in output, got:\n%s", out)
	}
}

func TestHumanRenderer_ResourceChanges(t *testing.T) {
	report := &diff.DiffReport{
		ChartName:   "mychart",
		OldVersion:  "1.0.0",
		NewVersion:  "2.0.0",
		GeneratedAt: time.Now(),
		Resources: []diff.ResourceDiff{
			{
				TemplateFile: "deployment.yaml",
				ResourceKind: "Deployment",
				ResourceName: "my-deploy",
				Changes:      []diff.Change{{Path: "image", Kind: diff.Changed, Risk: diff.RiskMedium}},
			},
			{
				TemplateFile: "new.yaml",
				ResourceKind: "Service",
				IsNew:        true,
			},
			{
				TemplateFile: "old.yaml",
				ResourceKind: "ConfigMap",
				IsRemoved:    true,
			},
		},
	}
	out := captureStdout(t, func() {
		r := &HumanRenderer{}
		if err := r.Render(report); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "NEW TEMPLATE") {
		t.Errorf("expected NEW TEMPLATE in output, got:\n%s", out)
	}
	if !strings.Contains(out, "TEMPLATE REMOVED") {
		t.Errorf("expected TEMPLATE REMOVED in output, got:\n%s", out)
	}
}

func TestFormatValue_Long(t *testing.T) {
	long := strings.Repeat("a", 100)
	result := formatValue(long)
	if len(result) > 80 {
		t.Errorf("expected truncation to <=80 chars, got %d", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Errorf("expected '...' suffix for long value, got %q", result)
	}
}

func TestFormatValue_Nil(t *testing.T) {
	if got := formatValue(nil); got != "<nil>" {
		t.Errorf("expected <nil>, got %q", got)
	}
}
