package diff

import (
	"time"

	"github.com/jerphil/helmdiff/internal/chart"
)

func Run(oldChart, newChart *chart.Chart) *DiffReport {
	report := &DiffReport{
		ChartName:   oldChart.Meta.Name,
		OldVersion:  oldChart.Meta.Version,
		NewVersion:  newChart.Meta.Version,
		GeneratedAt: time.Now(),
	}

	report.MetaChanges = ClassifyAll(DiffMeta(oldChart.Meta, newChart.Meta))
	report.ValueChanges = ClassifyAll(DiffValues(oldChart.Values, newChart.Values))
	report.CRDChanges = ClassifyAll(DiffCRDs(oldChart.CRDs, newChart.CRDs))

	resources := DiffTemplates(oldChart.Templates, newChart.Templates)
	for i := range resources {
		resources[i].Changes = ClassifyAll(resources[i].Changes)
	}
	report.Resources = resources

	return report
}
