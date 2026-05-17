package diff

import (
	"testing"

	"github.com/jerphil/helmdiff/internal/chart"
)

func TestDiffMeta_NoChanges(t *testing.T) {
	meta := chart.ChartMeta{AppVersion: "1.0", KubeVersion: ">=1.20.0", Type: "application"}
	changes := DiffMeta(meta, meta)
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestDiffMeta_AppVersionChanged(t *testing.T) {
	old := chart.ChartMeta{AppVersion: "1.9.5"}
	new := chart.ChartMeta{AppVersion: "1.11.0"}
	changes := DiffMeta(old, new)
	assertChange(t, changes, "appVersion", Changed, "1.9.5", "1.11.0")
}

func TestDiffMeta_KubeVersionChanged(t *testing.T) {
	old := chart.ChartMeta{KubeVersion: ">=1.20.0-0"}
	new := chart.ChartMeta{KubeVersion: ">=1.21.0-0"}
	changes := DiffMeta(old, new)
	assertChange(t, changes, "kubeVersion", Changed, ">=1.20.0-0", ">=1.21.0-0")
}

func TestDiffMeta_TypeChanged(t *testing.T) {
	old := chart.ChartMeta{Type: "application"}
	new := chart.ChartMeta{Type: "library"}
	changes := DiffMeta(old, new)
	assertChange(t, changes, "type", Changed, "application", "library")
}

func TestDiffMeta_DependencyAdded(t *testing.T) {
	old := chart.ChartMeta{}
	new := chart.ChartMeta{
		Dependencies: []chart.Dependency{{Name: "redis", Version: "1.0.0", Repository: "https://charts.example.com"}},
	}
	changes := DiffMeta(old, new)
	assertHasPath(t, changes, "dependencies.redis", Added)
}

func TestDiffMeta_DependencyRemoved(t *testing.T) {
	old := chart.ChartMeta{
		Dependencies: []chart.Dependency{{Name: "redis", Version: "1.0.0"}},
	}
	new := chart.ChartMeta{}
	changes := DiffMeta(old, new)
	assertHasPath(t, changes, "dependencies.redis", Removed)
}

func TestDiffMeta_DependencyVersionChanged(t *testing.T) {
	old := chart.ChartMeta{
		Dependencies: []chart.Dependency{{Name: "redis", Version: "1.0.0", Repository: "https://charts.example.com"}},
	}
	new := chart.ChartMeta{
		Dependencies: []chart.Dependency{{Name: "redis", Version: "2.0.0", Repository: "https://charts.example.com"}},
	}
	changes := DiffMeta(old, new)
	assertChange(t, changes, "dependencies.redis.version", Changed, "1.0.0", "2.0.0")
}

func TestDiffMeta_DependencyRepositoryChanged(t *testing.T) {
	old := chart.ChartMeta{
		Dependencies: []chart.Dependency{{Name: "redis", Version: "1.0.0", Repository: "https://old.example.com"}},
	}
	new := chart.ChartMeta{
		Dependencies: []chart.Dependency{{Name: "redis", Version: "1.0.0", Repository: "https://new.example.com"}},
	}
	changes := DiffMeta(old, new)
	assertChange(t, changes, "dependencies.redis.repository", Changed, "https://old.example.com", "https://new.example.com")
}

func TestCompareField_BothEmpty(t *testing.T) {
	changes := compareField("field", "", "")
	if len(changes) != 0 {
		t.Errorf("expected no changes for two empty strings, got %d", len(changes))
	}
}

func TestCompareField_OldEmpty(t *testing.T) {
	changes := compareField("field", "", "new")
	if len(changes) != 1 || changes[0].Kind != Added {
		t.Errorf("expected Added change, got %+v", changes)
	}
}

func TestCompareField_NewEmpty(t *testing.T) {
	changes := compareField("field", "old", "")
	if len(changes) != 1 || changes[0].Kind != Removed {
		t.Errorf("expected Removed change, got %+v", changes)
	}
}
