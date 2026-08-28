package audit

type Summary struct {
	Total  int            `json:"total"`
	ByKind map[string]int `json:"by_kind"`
}

func NewSummary() Summary { return Summary{ByKind: map[string]int{}} }

func (s *Summary) Add(kind string) {
	if kind == "" {
		kind = "unknown"
	}
	s.Total++
	s.ByKind[kind]++
}

func (s Summary) Count(kind string) int { return s.ByKind[kind] }
