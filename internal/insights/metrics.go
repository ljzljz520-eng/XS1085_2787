package insights

type Metrics struct {
	Candles     int     `json:"candles"`
	Reflections int     `json:"reflections"`
	Visitors    int     `json:"visitors"`
	QuietRate   float64 `json:"quiet_rate"`
}

func NewMetrics() Metrics { return Metrics{} }

func (m *Metrics) AddCandle()     { m.Candles++ }
func (m *Metrics) AddReflection() { m.Reflections++ }
func (m *Metrics) SetVisitors(count int) {
	if count < 0 {
		count = 0
	}
	m.Visitors = count
}

func (m *Metrics) CompleteSession() {
	total := m.Candles + m.Visitors
	if total == 0 {
		m.QuietRate = 1
		return
	}
	m.QuietRate = float64(m.Visitors) / float64(total)
}
