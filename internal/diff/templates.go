package diff

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jerphil/helmdiff/internal/chart"
	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
)

// helmDirective matches Go template directives: {{ ... }}
var helmDirective = regexp.MustCompile(`(?s)\{\{-?.*?-?\}\}`)

func DiffTemplates(oldTemplates, newTemplates []chart.Template) []ResourceDiff {
	var results []ResourceDiff

	oldMap := make(map[string]chart.Template)
	for _, t := range oldTemplates {
		oldMap[t.Name] = t
	}
	newMap := make(map[string]chart.Template)
	for _, t := range newTemplates {
		newMap[t.Name] = t
	}

	for name, nt := range newMap {
		kind := extractKind(nt.Content)
		if isTplHelper(name) {
			kind = "Helper"
		}
		ot, exists := oldMap[name]
		if !exists {
			results = append(results, ResourceDiff{
				TemplateFile: name,
				ResourceKind: kind,
				ResourceName: extractName(nt.Content),
				IsNew:        true,
			})
			continue
		}
		if changes := diffTemplate(name, ot.Content, nt.Content); len(changes) > 0 {
			results = append(results, ResourceDiff{
				TemplateFile: name,
				ResourceKind: kind,
				ResourceName: extractName(nt.Content),
				Changes:      changes,
			})
		}
	}

	// Templates only in old version (removed)
	for name, ot := range oldMap {
		if _, exists := newMap[name]; !exists {
			kind := extractKind(ot.Content)
			if isTplHelper(name) {
				kind = "Helper"
			}
			results = append(results, ResourceDiff{
				TemplateFile: name,
				ResourceKind: kind,
				ResourceName: extractName(ot.Content),
				IsRemoved:    true,
			})
		}
	}

	return results
}

func diffTemplate(name string, oldContent, newContent []byte) []Change {
	// Try semantic YAML diff after stripping template directives
	oldCleaned := stripHelmDirectives(oldContent)
	newCleaned := stripHelmDirectives(newContent)

	oldDocs := splitYAMLDocuments(oldCleaned)
	newDocs := splitYAMLDocuments(newCleaned)

	// Keep originals for readable fallback diffs
	oldOrigDocs := splitYAMLDocuments(oldContent)
	newOrigDocs := splitYAMLDocuments(newContent)

	var changes []Change
	maxDocs := len(oldDocs)
	if len(newDocs) > maxDocs {
		maxDocs = len(newDocs)
	}

	for i := 0; i < maxDocs; i++ {
		var oldDoc, newDoc []byte
		var oldOrig, newOrig []byte
		if i < len(oldDocs) {
			oldDoc = oldDocs[i]
		}
		if i < len(newDocs) {
			newDoc = newDocs[i]
		}
		if i < len(oldOrigDocs) {
			oldOrig = oldOrigDocs[i]
		}
		if i < len(newOrigDocs) {
			newOrig = newOrigDocs[i]
		}

		docChanges := diffYAMLDoc(i, oldDoc, newDoc, oldOrig, newOrig)
		changes = append(changes, docChanges...)
	}

	return changes
}

func diffYAMLDoc(idx int, oldDoc, newDoc, oldOrig, newOrig []byte) []Change {
	if len(bytes.TrimSpace(oldDoc)) == 0 && len(bytes.TrimSpace(newDoc)) == 0 {
		return nil
	}

	var oldMap, newMap map[string]any

	oldErr := yaml.Unmarshal(oldDoc, &oldMap)
	newErr := yaml.Unmarshal(newDoc, &newMap)

	if oldErr == nil && newErr == nil {
		prefix := ""
		if idx > 0 {
			prefix = fmt.Sprintf("doc[%d]", idx)
		}
		return diffMaps(prefix, normalizeMap(oldMap), normalizeMap(newMap))
	}

	// Fall back to unified diff using original content (not stripped) for readability
	if len(oldOrig) == 0 {
		oldOrig = oldDoc
	}
	if len(newOrig) == 0 {
		newOrig = newDoc
	}
	return lineDiff(oldOrig, newOrig)
}

// lineDiff falls back to unified diff using the original (non-stripped) content
// so the output is readable rather than full of "__helm__" placeholders.
func lineDiff(old, new []byte) []Change {
	if bytes.Equal(old, new) {
		return nil
	}
	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(old)),
		B:        difflib.SplitLines(string(new)),
		FromFile: "old",
		ToFile:   "new",
		Context:  3,
	})
	if strings.TrimSpace(diff) == "" {
		return nil
	}
	return []Change{{
		Path:     "(raw diff)",
		Kind:     Changed,
		OldValue: "(see diff)",
		NewValue: diff,
	}}
}

func isTplHelper(name string) bool {
	return strings.HasSuffix(name, ".tpl") ||
		strings.HasPrefix(filepath.Base(name), "_")
}

func stripHelmDirectives(content []byte) []byte {
	return helmDirective.ReplaceAll(content, []byte(`"__helm__"`))
}

func splitYAMLDocuments(content []byte) [][]byte {
	separator := []byte("\n---")
	docs := bytes.Split(content, separator)
	var result [][]byte
	for _, d := range docs {
		trimmed := bytes.TrimSpace(d)
		if len(trimmed) > 0 {
			result = append(result, trimmed)
		}
	}
	return result
}

func extractKind(content []byte) string {
	cleaned := stripHelmDirectives(content)
	var obj struct {
		Kind string `yaml:"kind"`
	}
	// Try each document
	for _, doc := range splitYAMLDocuments(cleaned) {
		if err := yaml.Unmarshal(doc, &obj); err == nil && obj.Kind != "" && obj.Kind != "__helm__" {
			return obj.Kind
		}
	}
	return "Unknown"
}

func extractName(content []byte) string {
	cleaned := stripHelmDirectives(content)
	var obj struct {
		Metadata struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
	}
	for _, doc := range splitYAMLDocuments(cleaned) {
		if err := yaml.Unmarshal(doc, &obj); err == nil && obj.Metadata.Name != "" && obj.Metadata.Name != "__helm__" {
			return obj.Metadata.Name
		}
	}
	return ""
}

// normalizeMap converts map[interface{}]interface{} (sometimes returned by yaml.v3) to map[string]any.
func normalizeMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = normalizeValue(v)
	}
	return result
}

func normalizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return normalizeMap(val)
	case map[any]any:
		m := make(map[string]any, len(val))
		for k, v2 := range val {
			m[fmt.Sprintf("%v", k)] = normalizeValue(v2)
		}
		return m
	case []any:
		for i, item := range val {
			val[i] = normalizeValue(item)
		}
		return val
	}
	return v
}
