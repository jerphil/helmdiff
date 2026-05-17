package diff

import (
	"fmt"

	"github.com/jerphil/helmdiff/internal/chart"
)

func DiffMeta(old, new chart.ChartMeta) []Change {
	var changes []Change

	changes = append(changes, compareField("appVersion", old.AppVersion, new.AppVersion)...)
	changes = append(changes, compareField("kubeVersion", old.KubeVersion, new.KubeVersion)...)
	changes = append(changes, compareField("type", old.Type, new.Type)...)
	changes = append(changes, diffDependencies(old.Dependencies, new.Dependencies)...)

	return changes
}

func compareField(path, old, new string) []Change {
	if old == new {
		return nil
	}
	if old == "" {
		return []Change{{Path: path, Kind: Added, NewValue: new}}
	}
	if new == "" {
		return []Change{{Path: path, Kind: Removed, OldValue: old}}
	}
	return []Change{{Path: path, Kind: Changed, OldValue: old, NewValue: new}}
}

func diffDependencies(old, new []chart.Dependency) []Change {
	var changes []Change

	oldMap := make(map[string]chart.Dependency)
	for _, d := range old {
		oldMap[d.Name] = d
	}
	newMap := make(map[string]chart.Dependency)
	for _, d := range new {
		newMap[d.Name] = d
	}

	for name, od := range oldMap {
		nd, exists := newMap[name]
		if !exists {
			changes = append(changes, Change{
				Path:     fmt.Sprintf("dependencies.%s", name),
				Kind:     Removed,
				OldValue: od.Version,
			})
			continue
		}
		if od.Version != nd.Version {
			changes = append(changes, Change{
				Path:     fmt.Sprintf("dependencies.%s.version", name),
				Kind:     Changed,
				OldValue: od.Version,
				NewValue: nd.Version,
			})
		}
		if od.Repository != nd.Repository {
			changes = append(changes, Change{
				Path:     fmt.Sprintf("dependencies.%s.repository", name),
				Kind:     Changed,
				OldValue: od.Repository,
				NewValue: nd.Repository,
			})
		}
	}

	for name, nd := range newMap {
		if _, exists := oldMap[name]; !exists {
			changes = append(changes, Change{
				Path:     fmt.Sprintf("dependencies.%s", name),
				Kind:     Added,
				NewValue: nd.Version,
			})
		}
	}

	return changes
}
