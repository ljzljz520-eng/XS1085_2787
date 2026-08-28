package insights

import "testing"

func TestReportSummarizesVisit(t *testing.T) {
	metrics := NewMetrics()
	metrics.AddCandle()
	metrics.AddReflection()
	metrics.SetVisitors(3)
	metrics.CompleteSession()
	report := BuildReport("纪念馆", metrics)
	if report.Metrics.Candles != 1 || report.Metrics.Reflections != 1 {
		t.Fatalf("unexpected metrics %#v", report.Metrics)
	}
	if len(report.Notes) != 3 || report.String() == "" {
		t.Fatalf("unexpected report %#v", report)
	}
}
