package diff

import (
	"testing"
)

func TestDiffValues_NoChanges(t *testing.T) {
	old := map[string]any{"replicas": 3, "image": map[string]any{"tag": "v1.0"}}
	got := DiffValues(old, old)
	if len(got) != 0 {
		t.Errorf("expected no changes, got %d", len(got))
	}
}

func TestDiffValues_AddedKey(t *testing.T) {
	old := map[string]any{"replicas": 1}
	new := map[string]any{"replicas": 1, "serviceAccount": map[string]any{"create": true}}
	changes := DiffValues(old, new)
	// A new top-level map key is reported as Added at the key level, not recursed into
	assertHasPath(t, changes, "serviceAccount", Added)
}

func TestDiffValues_RemovedKey(t *testing.T) {
	old := map[string]any{"replicas": 1, "podAnnotations": map[string]any{"foo": "bar"}}
	new := map[string]any{"replicas": 1}
	changes := DiffValues(old, new)
	assertHasPath(t, changes, "podAnnotations", Removed)
}

func TestDiffValues_ChangedLeaf(t *testing.T) {
	old := map[string]any{"image": map[string]any{"tag": "v1.0"}}
	new := map[string]any{"image": map[string]any{"tag": "v2.0"}}
	changes := DiffValues(old, new)
	assertChange(t, changes, "image.tag", Changed, "v1.0", "v2.0")
}

func TestDiffValues_NestedChange(t *testing.T) {
	old := map[string]any{
		"controller": map[string]any{
			"resources": map[string]any{
				"limits": map[string]any{"cpu": "100m"},
			},
		},
	}
	new := map[string]any{
		"controller": map[string]any{
			"resources": map[string]any{
				"limits": map[string]any{"cpu": "500m"},
			},
		},
	}
	changes := DiffValues(old, new)
	assertChange(t, changes, "controller.resources.limits.cpu", Changed, "100m", "500m")
}

func TestDiffValues_SliceIndexDiff(t *testing.T) {
	old := map[string]any{"ports": []any{80, 443}}
	new := map[string]any{"ports": []any{80, 443, 8080}}
	changes := DiffValues(old, new)
	assertChange(t, changes, "ports[2]", Added, nil, 8080)
}

func TestDiffValues_SliceItemRemoved(t *testing.T) {
	old := map[string]any{"ports": []any{80, 443}}
	new := map[string]any{"ports": []any{80}}
	changes := DiffValues(old, new)
	assertChange(t, changes, "ports[1]", Removed, 443, nil)
}

func TestDiffValues_EmptyMaps(t *testing.T) {
	changes := DiffValues(map[string]any{}, map[string]any{})
	if len(changes) != 0 {
		t.Errorf("expected no changes for two empty maps, got %d", len(changes))
	}
}

func TestDiffValues_NilOld(t *testing.T) {
	new := map[string]any{"key": "value"}
	changes := DiffValues(nil, new)
	assertHasPath(t, changes, "key", Added)
}

// helpers

func assertChange(t *testing.T, changes []Change, path string, kind ChangeKind, oldVal, newVal any) {
	t.Helper()
	for _, c := range changes {
		if c.Path == path && c.Kind == kind {
			if oldVal != nil && c.OldValue != oldVal {
				t.Errorf("path %q: OldValue = %v, want %v", path, c.OldValue, oldVal)
			}
			if newVal != nil && c.NewValue != newVal {
				t.Errorf("path %q: NewValue = %v, want %v", path, c.NewValue, newVal)
			}
			return
		}
	}
	t.Errorf("no %s change found for path %q in %+v", kind, path, changes)
}

func assertHasPath(t *testing.T, changes []Change, path string, kind ChangeKind) {
	t.Helper()
	for _, c := range changes {
		if c.Path == path && c.Kind == kind {
			return
		}
	}
	t.Errorf("no %s change found for path %q", kind, path)
}
