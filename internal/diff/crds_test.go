package diff

import (
	"testing"

	"github.com/jerphil/helmdiff/internal/chart"
)

func TestDiffCRDs_NoChanges(t *testing.T) {
	crds := []chart.CRD{{Name: "foo.example.com", Versions: []string{"v1"}}}
	changes := DiffCRDs(crds, crds)
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestDiffCRDs_Added(t *testing.T) {
	old := []chart.CRD{}
	new := []chart.CRD{{Name: "widgets.example.com", Group: "example.com", Kind: "Widget"}}
	changes := DiffCRDs(old, new)
	assertHasPath(t, changes, "crds.widgets.example.com", Added)
}

func TestDiffCRDs_Removed(t *testing.T) {
	old := []chart.CRD{{Name: "widgets.example.com"}}
	new := []chart.CRD{}
	changes := DiffCRDs(old, new)
	assertHasPath(t, changes, "crds.widgets.example.com", Removed)
}

func TestDiffCRDs_VersionAdded(t *testing.T) {
	old := []chart.CRD{{Name: "foo.example.com", Versions: []string{"v1"}}}
	new := []chart.CRD{{Name: "foo.example.com", Versions: []string{"v1", "v2"}}}
	changes := DiffCRDs(old, new)
	assertHasPath(t, changes, "crds.foo.example.com.versions", Added)
}

func TestDiffCRDs_VersionRemoved(t *testing.T) {
	old := []chart.CRD{{Name: "foo.example.com", Versions: []string{"v1", "v1beta1"}}}
	new := []chart.CRD{{Name: "foo.example.com", Versions: []string{"v1"}}}
	changes := DiffCRDs(old, new)
	assertHasPath(t, changes, "crds.foo.example.com.versions", Removed)
}

func TestDiffCRDs_KeyFallsBackToGroupKind(t *testing.T) {
	// CRDs without a Name field should key by group/kind
	old := []chart.CRD{{Group: "example.com", Kind: "Widget", Versions: []string{"v1"}}}
	new := []chart.CRD{}
	changes := DiffCRDs(old, new)
	if len(changes) == 0 {
		t.Error("expected a removed change for CRD keyed by group/kind")
	}
	if changes[0].Kind != Removed {
		t.Errorf("expected Removed, got %s", changes[0].Kind)
	}
}

func TestDiffCRDs_Multiple(t *testing.T) {
	old := []chart.CRD{
		{Name: "a.example.com"},
		{Name: "b.example.com"},
	}
	new := []chart.CRD{
		{Name: "b.example.com"},
		{Name: "c.example.com"},
	}
	changes := DiffCRDs(old, new)
	assertHasPath(t, changes, "crds.a.example.com", Removed)
	assertHasPath(t, changes, "crds.c.example.com", Added)
}
