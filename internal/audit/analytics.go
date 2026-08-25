package audit

import (
	"sort"
	"strings"
	"time"

	"memorialcandle/internal/memorial"
)

type Window struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type ActionStat struct {
	Action string `json:"action"`
	Count  int    `json:"count"`
}

type Timeline struct {
	Entries []memorial.AuditEntry `json:"entries"`
	Stats   []ActionStat          `json:"stats"`
	Window  Window                `json:"window"`
}

func NewWindow(entries []memorial.AuditEntry) Window {
	if len(entries) == 0 {
		return Window{}
	}
	start := entries[0].CreatedAt
	end := entries[0].CreatedAt
	for _, entry := range entries[1:] {
		if entry.CreatedAt.Before(start) {
			start = entry.CreatedAt
		}
		if entry.CreatedAt.After(end) {
			end = entry.CreatedAt
		}
	}
	return Window{Start: start, End: end}
}

func BuildTimeline(entries []memorial.AuditEntry) Timeline {
	ordered := append([]memorial.AuditEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
	})
	counts := map[string]int{}
	for _, entry := range ordered {
		action := strings.TrimSpace(entry.Action)
		if action == "" {
			action = "unknown"
		}
		counts[action]++
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	stats := make([]ActionStat, 0, len(keys))
	for _, key := range keys {
		stats = append(stats, ActionStat{Action: key, Count: counts[key]})
	}
	return Timeline{Entries: ordered, Stats: stats, Window: NewWindow(ordered)}
}

func (t Timeline) Valid() bool {
	if len(t.Entries) == 0 {
		return true
	}
	return !t.Window.Start.IsZero() && !t.Window.End.IsZero() && !t.Window.End.Before(t.Window.Start)
}

func (t Timeline) Count(action string) int {
	for _, stat := range t.Stats {
		if stat.Action == action {
			return stat.Count
		}
	}
	return 0
}

func Filter(entries []memorial.AuditEntry, action, subject string) []memorial.AuditEntry {
	result := make([]memorial.AuditEntry, 0, len(entries))
	for _, entry := range entries {
		if action != "" && entry.Action != action {
			continue
		}
		if subject != "" && entry.Subject != subject {
			continue
		}
		result = append(result, entry)
	}
	return result
}

func Last(entries []memorial.AuditEntry, action string) (memorial.AuditEntry, bool) {
	filtered := Filter(entries, action, "")
	if len(filtered) == 0 {
		return memorial.AuditEntry{}, false
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	return filtered[0], true
}

func QuietRatio(entries []memorial.AuditEntry) float64 {
	if len(entries) == 0 {
		return 1
	}
	quiet := 0
	for _, entry := range entries {
		if entry.Action == "close" || entry.Action == "return" {
			quiet++
		}
	}
	return float64(quiet) / float64(len(entries))
}

func Merge(left, right Timeline) Timeline {
	entries := make([]memorial.AuditEntry, 0, len(left.Entries)+len(right.Entries))
	entries = append(entries, left.Entries...)
	entries = append(entries, right.Entries...)
	return BuildTimeline(entries)
}
