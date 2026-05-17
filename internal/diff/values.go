package diff

import "fmt"

// DiffValues performs a deep semantic diff of two parsed YAML value maps.
func DiffValues(oldVals, newVals map[string]any) []Change {
	return diffMaps("", oldVals, newVals)
}

func diffMaps(prefix string, old, new map[string]any) []Change {
	var changes []Change

	for k, oldVal := range old {
		path := joinPath(prefix, k)
		newVal, exists := new[k]
		if !exists {
			changes = append(changes, Change{
				Path:     path,
				Kind:     Removed,
				OldValue: oldVal,
			})
			continue
		}
		changes = append(changes, diffValues(path, oldVal, newVal)...)
	}

	for k, newVal := range new {
		if _, exists := old[k]; !exists {
			changes = append(changes, Change{
				Path:     joinPath(prefix, k),
				Kind:     Added,
				NewValue: newVal,
			})
		}
	}

	return changes
}

func diffValues(path string, old, new any) []Change {
	// Both maps: recurse
	oldMap, oldIsMap := toStringMap(old)
	newMap, newIsMap := toStringMap(new)
	if oldIsMap && newIsMap {
		return diffMaps(path, oldMap, newMap)
	}

	// Both slices: diff by index
	oldSlice, oldIsSlice := toSlice(old)
	newSlice, newIsSlice := toSlice(new)
	if oldIsSlice && newIsSlice {
		return diffSlices(path, oldSlice, newSlice)
	}

	// Leaf comparison
	if fmt.Sprintf("%v", old) != fmt.Sprintf("%v", new) {
		return []Change{{
			Path:     path,
			Kind:     Changed,
			OldValue: old,
			NewValue: new,
		}}
	}
	return nil
}

func diffSlices(path string, old, new []any) []Change {
	var changes []Change
	max := len(old)
	if len(new) > max {
		max = len(new)
	}
	for i := 0; i < max; i++ {
		iPath := fmt.Sprintf("%s[%d]", path, i)
		if i >= len(old) {
			changes = append(changes, Change{Path: iPath, Kind: Added, NewValue: new[i]})
		} else if i >= len(new) {
			changes = append(changes, Change{Path: iPath, Kind: Removed, OldValue: old[i]})
		} else {
			changes = append(changes, diffValues(iPath, old[i], new[i])...)
		}
	}
	return changes
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func toStringMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func toSlice(v any) ([]any, bool) {
	s, ok := v.([]any)
	return s, ok
}
