package renderer

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jerphil/helmdiff/internal/diff"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestJSONRenderer_ValidJSON(t *testing.T) {
	report := &diff.DiffReport{
		ChartName:  "mychart",
		OldVersion: "1.0.0",
		NewVersion: "2.0.0",
		GeneratedAt: time.Now(),
		MetaChanges: []diff.Change{
			{Path: "appVersion", Kind: diff.Changed, OldValue: "1.0", NewValue: "2.0", Risk: diff.RiskMedium},
		},
	}

	out := captureStdout(t, func() {
		r := &JSONRenderer{}
		if err := r.Render(report); err != nil {
			t.Fatalf("Render failed: %v", err)
		}
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	if parsed["ChartName"] != "mychart" {
		t.Errorf("expected ChartName=mychart, got %v", parsed["ChartName"])
	}
}

func TestJSONRenderer_EmptyReport(t *testing.T) {
	report := &diff.DiffReport{
		ChartName:   "empty",
		OldVersion:  "1.0.0",
		NewVersion:  "1.0.1",
		GeneratedAt: time.Now(),
	}

	out := captureStdout(t, func() {
		r := &JSONRenderer{}
		if err := r.Render(report); err != nil {
			t.Fatalf("Render failed: %v", err)
		}
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}
