package cmd

import (
	"testing"

	"github.com/jerphil/helmdiff/internal/diff"
)

func TestParseRiskLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected diff.RiskLevel
		wantErr  bool
	}{
		{"low", diff.RiskLow, false},
		{"medium", diff.RiskMedium, false},
		{"high", diff.RiskHigh, false},
		{"critical", diff.RiskCritical, false},
		{"LOW", diff.RiskLow, false},
		{"HIGH", diff.RiskHigh, false},
		{"Critical", diff.RiskCritical, false},
		{"unknown", diff.RiskLow, true},
		{"", diff.RiskLow, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseRiskLevel(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
