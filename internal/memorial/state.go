package memorial

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

type CandleState string

const (
	CandleQuiet     CandleState = "quiet"
	CandleSpreading CandleState = "spreading"
	CandleClosed    CandleState = "closed"
)

type AnimationStatus string

const (
	AnimationPending  AnimationStatus = "pending"
	AnimationRunning  AnimationStatus = "running"
	AnimationStopped  AnimationStatus = "stopped"
	AnimationFinished AnimationStatus = "finished"
)

type AnimationJob struct {
	ID        string          `json:"id"`
	CandleID  string          `json:"candle_id"`
	Status    AnimationStatus `json:"status"`
	Generated int             `json:"generated"`
	StartedAt time.Time       `json:"started_at"`
	StoppedAt time.Time       `json:"stopped_at"`
}

type Session struct {
	ID         string    `json:"id"`
	MemorialID string    `json:"memorial_id"`
	VisitorID  string    `json:"visitor_id"`
	Angle      float64   `json:"angle"`
	Zoom       float64   `json:"zoom"`
	Active     bool      `json:"active"`
	StartedAt  time.Time `json:"started_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CandleSnapshot struct {
	Candle     Candle         `json:"candle"`
	Job        AnimationJob   `json:"job"`
	Session    VisitorSession `json:"session"`
	SparkStats SparkStats     `json:"spark_stats"`
	Quiet      bool           `json:"quiet"`
}

type SparkStats struct {
	Count        int     `json:"count"`
	AverageGlow  float64 `json:"average_glow"`
	MinRadius    float64 `json:"min_radius"`
	MaxRadius    float64 `json:"max_radius"`
	LastSequence int     `json:"last_sequence"`
}

func NewAnimationJob(id, candleID string, at time.Time) AnimationJob {
	return AnimationJob{ID: id, CandleID: candleID, Status: AnimationPending, StartedAt: at}
}

func (j AnimationJob) Valid() bool {
	return j.ID != "" && j.CandleID != "" && j.StartedAt.IsZero() == false && j.Generated >= 0
}

func (j *AnimationJob) Start(at time.Time) error {
	if j == nil || j.ID == "" || j.CandleID == "" {
		return errors.New("animation job identity is required")
	}
	if j.Status != AnimationPending {
		return errors.New("animation job cannot start from current status")
	}
	j.Status = AnimationRunning
	j.StartedAt = at
	return nil
}

func (j *AnimationJob) AddSpark() error {
	if j == nil {
		return errors.New("animation job is nil")
	}
	if j.Status != AnimationRunning {
		return errors.New("animation job is not running")
	}
	j.Generated++
	return nil
}

func (j *AnimationJob) Stop(at time.Time) error {
	if j == nil {
		return errors.New("animation job is nil")
	}
	if j.Status == AnimationStopped || j.Status == AnimationFinished {
		return nil
	}
	if j.Status != AnimationRunning && j.Status != AnimationPending {
		return errors.New("animation job cannot stop from current status")
	}
	j.Status = AnimationStopped
	j.StoppedAt = at
	return nil
}

func (j *AnimationJob) Finish(at time.Time) error {
	if j == nil {
		return errors.New("animation job is nil")
	}
	if j.Status != AnimationRunning {
		return errors.New("animation job is not running")
	}
	j.Status = AnimationFinished
	j.StoppedAt = at
	return nil
}

func (c Candle) State() CandleState {
	if c.Extinguished {
		return CandleClosed
	}
	return CandleQuiet
}

func (c *Candle) BeginSpread() error {
	if c == nil {
		return errors.New("candle is nil")
	}
	if c.Extinguished {
		return errors.New("closed candle cannot spread")
	}
	return nil
}

func (c *Candle) Extinguish() error {
	if c == nil {
		return errors.New("candle is nil")
	}
	if c.Extinguished {
		return nil
	}
	c.Extinguished = true
	return nil
}

func (c Candle) Quiet() bool {
	return c.Extinguished || c.Intensity <= 0
}

func (c Candle) NormalizedMessage() string {
	return strings.Join(strings.Fields(c.Message), " ")
}

func (c Candle) ValidMessage() bool {
	message := c.NormalizedMessage()
	return len([]rune(message)) >= 1 && len([]rune(message)) <= 240
}

func (s VisitorSession) AsSession() Session {
	return Session{ID: s.ID, MemorialID: s.MemorialID, VisitorID: s.ID, Angle: s.Angle, Zoom: s.Zoom, Active: s.Active, StartedAt: s.StartedAt, UpdatedAt: s.LastSeenAt}
}

func (s Session) Valid() bool {
	return s.ID != "" && s.MemorialID != "" && s.VisitorID != "" && s.Zoom >= 0.75 && s.Zoom <= 3 && s.Angle >= 0 && s.Angle < 360
}

func (s *Session) Touch(at time.Time) error {
	if s == nil || s.ID == "" {
		return errors.New("session identity is required")
	}
	if !s.Active {
		return ErrClosed
	}
	s.UpdatedAt = at
	return nil
}

func NormalizeAngle(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func AngleDistance(from, to float64) float64 {
	from = NormalizeAngle(from)
	to = NormalizeAngle(to)
	difference := math.Abs(to - from)
	if difference > 180 {
		return 360 - difference
	}
	return difference
}

func BuildSparkStats(sparks []Spark) SparkStats {
	stats := SparkStats{}
	if len(sparks) == 0 {
		return stats
	}
	stats.Count = len(sparks)
	stats.MinRadius = sparks[0].Radius
	stats.MaxRadius = sparks[0].Radius
	for _, spark := range sparks {
		stats.AverageGlow += spark.Brightness
		if spark.Radius < stats.MinRadius {
			stats.MinRadius = spark.Radius
		}
		if spark.Radius > stats.MaxRadius {
			stats.MaxRadius = spark.Radius
		}
		if spark.Sequence > stats.LastSequence {
			stats.LastSequence = spark.Sequence
		}
	}
	stats.AverageGlow /= float64(stats.Count)
	return stats
}

func StableSparkWindow(sparks []Spark, limit int) []Spark {
	copyOf := append([]Spark(nil), sparks...)
	sort.SliceStable(copyOf, func(i, j int) bool {
		if copyOf[i].Sequence == copyOf[j].Sequence {
			return copyOf[i].ID < copyOf[j].ID
		}
		return copyOf[i].Sequence > copyOf[j].Sequence
	})
	if limit < 0 {
		limit = 0
	}
	if len(copyOf) > limit {
		copyOf = copyOf[:limit]
	}
	return copyOf
}

func Transition(from CandleState, event string) (CandleState, error) {
	switch from {
	case CandleQuiet:
		if event == "spread" {
			return CandleSpreading, nil
		}
		if event == "close" {
			return CandleClosed, nil
		}
	case CandleSpreading:
		if event == "settle" {
			return CandleQuiet, nil
		}
		if event == "close" {
			return CandleClosed, nil
		}
	case CandleClosed:
		if event == "close" {
			return CandleClosed, nil
		}
	}
	return from, errors.New("invalid candle state transition")
}

func QuietFrame(job AnimationJob, candle Candle) string {
	if candle.Extinguished || job.Status == AnimationStopped {
		return "quiet"
	}
	if job.Status == AnimationRunning {
		return "spreading"
	}
	return "ready"
}
