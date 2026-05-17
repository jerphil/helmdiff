package diff

import (
	"fmt"
	"strings"
)

type rule struct {
	match    func(Change) bool
	risk     RiskLevel
	describe func(Change) string
}

var rules = []rule{
	{
		match: func(c Change) bool {
			return strings.HasPrefix(c.Path, "crds.") && c.Kind == Removed
		},
		risk:     RiskCritical,
		describe: func(c Change) string { return fmt.Sprintf("CRD removed: %s", c.OldValue) },
	},
	{
		match: func(c Change) bool {
			return strings.HasPrefix(c.Path, "crds.") && c.Kind == Added
		},
		risk:     RiskHigh,
		describe: func(c Change) string { return fmt.Sprintf("CRD added: %s", c.NewValue) },
	},
	{
		match: func(c Change) bool {
			return strings.HasPrefix(c.Path, "dependencies.") && (c.Kind == Added || c.Kind == Removed)
		},
		risk:     RiskHigh,
		describe: func(c Change) string { return fmt.Sprintf("dependency %s: %s", c.Kind, c.Path) },
	},
	{
		match: func(c Change) bool {
			return containsAny(c.Path, "resources.limits", "resources.requests")
		},
		risk:     RiskHigh,
		describe: resourceDesc,
	},
	{
		match: func(c Change) bool {
			return containsAny(c.Path, "securityContext", "podSecurityContext")
		},
		risk:     RiskHigh,
		describe: classifyGenericDesc("security context"),
	},
	{
		match: func(c Change) bool {
			return containsAny(c.Path, "rbac", "ClusterRole", "serviceAccount", "clusterrole")
		},
		risk:     RiskHigh,
		describe: classifyGenericDesc("RBAC/serviceAccount"),
	},
	{
		match: func(c Change) bool {
			return containsAny(c.Path, "image.tag", "image.repository", "image.registry")
		},
		risk:     RiskMedium,
		describe: imageDesc,
	},
	{
		match: func(c Change) bool {
			return containsAny(c.Path, "ingress.enabled", "service.type", "service.port")
		},
		risk:     RiskMedium,
		describe: classifyGenericDesc("network/ingress"),
	},
	{
		match: func(c Change) bool {
			return strings.EqualFold(c.Path, "appVersion") || strings.EqualFold(c.Path, "kubeVersion")
		},
		risk: RiskMedium,
		describe: func(c Change) string {
			return fmt.Sprintf("%s changed: %v → %v", c.Path, c.OldValue, c.NewValue)
		},
	},
	{
		match: func(c Change) bool {
			return containsAny(c.Path, "annotations", "labels")
		},
		risk:     RiskLow,
		describe: classifyGenericDesc("annotation/label"),
	},
	{
		match:    func(c Change) bool { return c.Path == "(raw diff)" },
		risk:     RiskMedium,
		describe: func(c Change) string { return "template structure changed (complex template)" },
	},
}

// Classify sets Risk and Description on a Change based on ordered rules.
func Classify(c Change) Change {
	for _, r := range rules {
		if r.match(c) {
			c.Risk = r.risk
			if c.Description == "" {
				c.Description = r.describe(c)
			}
			return c
		}
	}
	c.Risk = RiskLow
	if c.Description == "" {
		c.Description = classifyDefaultDesc(c)
	}
	return c
}

// ClassifyAll classifies a slice of changes in place.
func ClassifyAll(changes []Change) []Change {
	for i, c := range changes {
		changes[i] = Classify(c)
	}
	return changes
}

func containsAny(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func resourceDesc(c Change) string {
	return fmt.Sprintf("resource %s changed: %v → %v", c.Path, c.OldValue, c.NewValue)
}

func imageDesc(c Change) string {
	return fmt.Sprintf("image %s changed: %v → %v", c.Path, c.OldValue, c.NewValue)
}

func classifyGenericDesc(label string) func(Change) string {
	return func(c Change) string {
		switch c.Kind {
		case Added:
			return fmt.Sprintf("%s added at %s: %v", label, c.Path, c.NewValue)
		case Removed:
			return fmt.Sprintf("%s removed at %s (was: %v)", label, c.Path, c.OldValue)
		default:
			return fmt.Sprintf("%s changed at %s: %v → %v", label, c.Path, c.OldValue, c.NewValue)
		}
	}
}

func classifyDefaultDesc(c Change) string {
	switch c.Kind {
	case Added:
		return fmt.Sprintf("added %s: %v", c.Path, c.NewValue)
	case Removed:
		return fmt.Sprintf("removed %s (was: %v)", c.Path, c.OldValue)
	default:
		return fmt.Sprintf("%s changed: %v → %v", c.Path, c.OldValue, c.NewValue)
	}
}
