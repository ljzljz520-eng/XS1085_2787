package insights

import (
	"fmt"
	"sort"
	"strings"
)

type Dashboard struct {
	Title      string       `json:"title"`
	Headline   string       `json:"headline"`
	Report     Report       `json:"report"`
	Highlights []Highlight  `json:"highlights"`
	Trend      []TrendPoint `json:"trend"`
	Quiet      bool         `json:"quiet"`
}

type Highlight struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Tone  string `json:"tone"`
}

type TrendPoint struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type EventSample struct {
	Kind  string
	Value int
}

func BuildDashboard(title string, metrics Metrics, samples []EventSample) Dashboard {
	report := BuildReport(title, metrics)
	headline := "烛光仍在"
	if metrics.QuietRate >= 0.75 {
		headline = "大多数记忆回到安静"
	}
	if metrics.Candles == 0 {
		headline = "等待第一盏烛光"
	}
	highlights := []Highlight{
		{Label: "烛光", Value: formatCount(metrics.Candles), Tone: "amber"},
		{Label: "访客", Value: formatCount(metrics.Visitors), Tone: "blue"},
		{Label: "留言", Value: formatCount(metrics.Reflections), Tone: "rose"},
		{Label: "安静率", Value: fmt.Sprintf("%.0f%%", metrics.QuietRate*100), Tone: "slate"},
	}
	return Dashboard{Title: title, Headline: headline, Report: report, Highlights: highlights, Trend: Trend(samples), Quiet: metrics.QuietRate >= 0.5}
}

func formatCount(value int) string {
	if value < 0 {
		value = 0
	}
	return fmt.Sprintf("%d", value)
}

func Trend(samples []EventSample) []TrendPoint {
	grouped := map[string]int{}
	for _, sample := range samples {
		key := strings.TrimSpace(sample.Kind)
		if key == "" {
			key = "other"
		}
		if sample.Value < 0 {
			sample.Value = 0
		}
		grouped[key] += sample.Value
	}
	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]TrendPoint, 0, len(keys))
	for _, key := range keys {
		result = append(result, TrendPoint{Label: key, Value: grouped[key]})
	}
	return result
}

func TrendPeak(points []TrendPoint) (TrendPoint, bool) {
	if len(points) == 0 {
		return TrendPoint{}, false
	}
	peak := points[0]
	for _, point := range points[1:] {
		if point.Value > peak.Value || point.Value == peak.Value && point.Label < peak.Label {
			peak = point
		}
	}
	return peak, true
}

func (d Dashboard) Highlight(label string) (Highlight, bool) {
	for _, highlight := range d.Highlights {
		if highlight.Label == label {
			return highlight, true
		}
	}
	return Highlight{}, false
}

func (d Dashboard) Summary() string {
	peak, ok := TrendPeak(d.Trend)
	if !ok {
		return d.Headline
	}
	return fmt.Sprintf("%s；高峰为%s(%d)", d.Headline, peak.Label, peak.Value)
}

func (d Dashboard) Valid() bool {
	return strings.TrimSpace(d.Title) != "" && strings.TrimSpace(d.Headline) != "" && len(d.Highlights) >= 3 && d.Report.Title == d.Title
}

func (m Metrics) Add(sample EventSample) Metrics {
	if sample.Value < 0 {
		sample.Value = 0
	}
	switch sample.Kind {
	case "candle":
		m.Candles += sample.Value
	case "reflection":
		m.Reflections += sample.Value
	case "visitor":
		m.Visitors += sample.Value
	}
	return m
}

func (m Metrics) Merge(other Metrics) Metrics {
	return Metrics{Candles: m.Candles + other.Candles, Reflections: m.Reflections + other.Reflections, Visitors: m.Visitors + other.Visitors, QuietRate: (m.QuietRate + other.QuietRate) / 2}
}

func (m Metrics) Empty() bool {
	return m.Candles == 0 && m.Reflections == 0 && m.Visitors == 0
}

func NormalizeRate(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
