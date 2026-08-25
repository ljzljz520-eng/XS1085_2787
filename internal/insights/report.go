package insights

import (
	"fmt"
	"sort"
)

type Report struct {
	Title   string   `json:"title"`
	Metrics Metrics  `json:"metrics"`
	Notes   []string `json:"notes"`
}

func BuildReport(title string, metrics Metrics) Report {
	notes := []string{}
	if metrics.Candles > 0 {
		notes = append(notes, "烛光正在被记住")
	}
	if metrics.Reflections > 0 {
		notes = append(notes, "有人留下了思念")
	}
	if metrics.QuietRate >= 0.5 {
		notes = append(notes, "大多数访客回到了安静")
	}
	sort.Strings(notes)
	return Report{Title: title, Metrics: metrics, Notes: notes}
}

func (r Report) String() string {
	return fmt.Sprintf("%s: %d candles, %d reflections", r.Title, r.Metrics.Candles, r.Metrics.Reflections)
}
