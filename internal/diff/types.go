package diff

import "time"

type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
	RiskCritical
)

func (r RiskLevel) String() string {
	switch r {
	case RiskCritical:
		return "CRITICAL"
	case RiskHigh:
		return "HIGH"
	case RiskMedium:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

type ChangeKind string

const (
	Added   ChangeKind = "added"
	Removed ChangeKind = "removed"
	Changed ChangeKind = "changed"
)

type Change struct {
	Path        string
	Kind        ChangeKind
	OldValue    any
	NewValue    any
	Risk        RiskLevel
	Description string
}

type ResourceDiff struct {
	TemplateFile string
	ResourceKind string
	ResourceName string
	Changes      []Change
	IsNew        bool // template only in new version
	IsRemoved    bool // template only in old version
}

type DiffReport struct {
	ChartName    string
	OldVersion   string
	NewVersion   string
	MetaChanges  []Change
	ValueChanges []Change
	Resources    []ResourceDiff
	CRDChanges   []Change
	GeneratedAt  time.Time
}

// MaxRisk returns the highest risk level found across all changes.
func (r *DiffReport) MaxRisk() RiskLevel {
	max := RiskLow
	all := append(append(r.MetaChanges, r.ValueChanges...), r.CRDChanges...)
	for _, c := range all {
		if c.Risk > max {
			max = c.Risk
		}
	}
	for _, res := range r.Resources {
		for _, c := range res.Changes {
			if c.Risk > max {
				max = c.Risk
			}
		}
	}
	return max
}

func (r *DiffReport) HighCount() int {
	return r.countByRisk(RiskHigh, RiskCritical)
}

func (r *DiffReport) MediumCount() int {
	return r.countByRisk(RiskMedium)
}

func (r *DiffReport) LowCount() int {
	return r.countByRisk(RiskLow)
}

func (r *DiffReport) countByRisk(levels ...RiskLevel) int {
	set := make(map[RiskLevel]bool)
	for _, l := range levels {
		set[l] = true
	}
	count := 0
	all := append(append(r.MetaChanges, r.ValueChanges...), r.CRDChanges...)
	for _, c := range all {
		if set[c.Risk] {
			count++
		}
	}
	for _, res := range r.Resources {
		for _, c := range res.Changes {
			if set[c.Risk] {
				count++
			}
		}
	}
	return count
}
