package diff

import (
	"fmt"

	"github.com/jerphil/helmdiff/internal/chart"
)

func DiffCRDs(old, new []chart.CRD) []Change {
	var changes []Change

	oldMap := make(map[string]chart.CRD)
	for _, c := range old {
		key := crdKey(c)
		oldMap[key] = c
	}
	newMap := make(map[string]chart.CRD)
	for _, c := range new {
		key := crdKey(c)
		newMap[key] = c
	}

	for key, oc := range oldMap {
		nc, exists := newMap[key]
		if !exists {
			changes = append(changes, Change{
				Path:     fmt.Sprintf("crds.%s", oc.Name),
				Kind:     Removed,
				OldValue: oc.Name,
			})
			continue
		}
		// Check for version changes
		oldVers := versionsSet(oc.Versions)
		newVers := versionsSet(nc.Versions)
		for v := range oldVers {
			if !newVers[v] {
				changes = append(changes, Change{
					Path:     fmt.Sprintf("crds.%s.versions", oc.Name),
					Kind:     Removed,
					OldValue: v,
				})
			}
		}
		for v := range newVers {
			if !oldVers[v] {
				changes = append(changes, Change{
					Path:     fmt.Sprintf("crds.%s.versions", nc.Name),
					Kind:     Added,
					NewValue: v,
				})
			}
		}
	}

	for key, nc := range newMap {
		if _, exists := oldMap[key]; !exists {
			changes = append(changes, Change{
				Path:     fmt.Sprintf("crds.%s", nc.Name),
				Kind:     Added,
				NewValue: nc.Name,
			})
		}
	}

	return changes
}

func crdKey(c chart.CRD) string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("%s/%s", c.Group, c.Kind)
}

func versionsSet(vs []string) map[string]bool {
	m := make(map[string]bool, len(vs))
	for _, v := range vs {
		m[v] = true
	}
	return m
}
