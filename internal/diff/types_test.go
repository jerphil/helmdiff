package diff

import "testing"

func TestRiskLevelString(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected string
	}{
		{RiskLow, "LOW"},
		{RiskMedium, "MEDIUM"},
		{RiskHigh, "HIGH"},
		{RiskCritical, "CRITICAL"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("RiskLevel(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func makeReport() *DiffReport {
	return &DiffReport{
		MetaChanges:  []Change{{Risk: RiskLow}, {Risk: RiskHigh}},
		ValueChanges: []Change{{Risk: RiskMedium}},
		CRDChanges:   []Change{{Risk: RiskCritical}},
		Resources: []ResourceDiff{
			{Changes: []Change{{Risk: RiskHigh}, {Risk: RiskMedium}}},
		},
	}
}

func TestHighCount(t *testing.T) {
	r := makeReport()
	// HIGH: 2 (MetaChanges + Resources), CRITICAL: 1 (CRDChanges) → total 3
	if got := r.HighCount(); got != 3 {
		t.Errorf("HighCount() = %d, want 3", got)
	}
}

func TestMediumCount(t *testing.T) {
	r := makeReport()
	// MEDIUM: 1 (ValueChanges) + 1 (Resources) = 2
	if got := r.MediumCount(); got != 2 {
		t.Errorf("MediumCount() = %d, want 2", got)
	}
}

func TestLowCount(t *testing.T) {
	r := makeReport()
	// LOW: 1 (MetaChanges)
	if got := r.LowCount(); got != 1 {
		t.Errorf("LowCount() = %d, want 1", got)
	}
}

func TestCountByRisk_Empty(t *testing.T) {
	r := &DiffReport{}
	if r.HighCount() != 0 || r.MediumCount() != 0 || r.LowCount() != 0 {
		t.Error("expected all counts to be 0 for empty report")
	}
}

func TestMaxRisk_Empty(t *testing.T) {
	r := &DiffReport{}
	if r.MaxRisk() != RiskLow {
		t.Errorf("expected RiskLow for empty report, got %s", r.MaxRisk())
	}
}

func TestMaxRisk_SingleHigh(t *testing.T) {
	r := &DiffReport{
		MetaChanges: []Change{{Risk: RiskHigh}},
	}
	if r.MaxRisk() != RiskHigh {
		t.Errorf("expected RiskHigh, got %s", r.MaxRisk())
	}
}

func TestMaxRisk_CriticalWins(t *testing.T) {
	r := &DiffReport{
		MetaChanges:  []Change{{Risk: RiskHigh}},
		ValueChanges: []Change{{Risk: RiskLow}},
		CRDChanges:   []Change{{Risk: RiskCritical}},
	}
	if r.MaxRisk() != RiskCritical {
		t.Errorf("expected RiskCritical, got %s", r.MaxRisk())
	}
}

func TestMaxRisk_FromResources(t *testing.T) {
	r := &DiffReport{
		Resources: []ResourceDiff{
			{Changes: []Change{{Risk: RiskMedium}, {Risk: RiskHigh}}},
		},
	}
	if r.MaxRisk() != RiskHigh {
		t.Errorf("expected RiskHigh, got %s", r.MaxRisk())
	}
}
