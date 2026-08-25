package memorial

import "time"

type Memorial struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Dedication  string    `json:"dedication"`
	CreatedAt   time.Time `json:"created_at"`
	CandleCount int       `json:"candle_count"`
	Quiet       bool      `json:"quiet"`
}

type Candle struct {
	ID           string    `json:"id"`
	MemorialID   string    `json:"memorial_id"`
	VisitorID    string    `json:"visitor_id"`
	Message      string    `json:"message"`
	Color        string    `json:"color"`
	Intensity    int       `json:"intensity"`
	LitAt        time.Time `json:"lit_at"`
	Extinguished bool      `json:"extinguished"`
}

type Spark struct {
	ID         string    `json:"id"`
	CandleID   string    `json:"candle_id"`
	Sequence   int       `json:"sequence"`
	Angle      float64   `json:"angle"`
	Radius     float64   `json:"radius"`
	Brightness float64   `json:"brightness"`
	CreatedAt  time.Time `json:"created_at"`
}

type VisitorSession struct {
	ID         string    `json:"id"`
	MemorialID string    `json:"memorial_id"`
	StartedAt  time.Time `json:"started_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	Angle      float64   `json:"angle"`
	Zoom       float64   `json:"zoom"`
	Active     bool      `json:"active"`
}

type Reflection struct {
	ID         string    `json:"id"`
	MemorialID string    `json:"memorial_id"`
	VisitorID  string    `json:"visitor_id"`
	Text       string    `json:"text"`
	Moderated  bool      `json:"moderated"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditEntry struct {
	ID         string    `json:"id"`
	MemorialID string    `json:"memorial_id"`
	Action     string    `json:"action"`
	Subject    string    `json:"subject"`
	Detail     string    `json:"detail"`
	CreatedAt  time.Time `json:"created_at"`
}

type Scene struct {
	Memorial    Memorial     `json:"memorial"`
	Candle      Candle       `json:"candle"`
	Sparks      []Spark      `json:"sparks"`
	Visitors    int          `json:"visitors"`
	Reflections []Reflection `json:"reflections"`
}

type SceneEvent struct {
	Kind      string `json:"kind"`
	SessionID string `json:"session_id"`
	CandleID  string `json:"candle_id"`
	Spark     *Spark `json:"spark,omitempty"`
	Quiet     bool   `json:"quiet"`
}

func (m Memorial) Valid() bool {
	return m.ID != "" && m.Title != "" && m.Dedication != "" && m.CandleCount >= 0
}

func (c Candle) Valid() bool {
	return c.ID != "" && c.MemorialID != "" && c.VisitorID != "" && c.Intensity >= 1 && c.Intensity <= 10
}

func (s VisitorSession) Valid() bool {
	return s.ID != "" && s.MemorialID != "" && s.Zoom > 0 && s.Zoom <= 4
}

func (r Reflection) Valid() bool {
	return r.ID != "" && r.MemorialID != "" && r.VisitorID != "" && len(r.Text) >= 2 && len(r.Text) <= 280
}
