package memorial

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type OrbitPlan struct {
	CandleID string  `json:"candle_id"`
	Count    int     `json:"count"`
	Radius   float64 `json:"radius"`
	Tilt     float64 `json:"tilt"`
	Speed    float64 `json:"speed"`
}

type OrbitFrame struct {
	Index      int     `json:"index"`
	Angle      float64 `json:"angle"`
	Radius     float64 `json:"radius"`
	Brightness float64 `json:"brightness"`
	Opacity    float64 `json:"opacity"`
}

type CeremonyCue struct {
	At       int    `json:"at"`
	Kind     string `json:"kind"`
	Message  string `json:"message"`
	Duration int    `json:"duration"`
}

type ScenePlan struct {
	Title     string        `json:"title"`
	Subtitle  string        `json:"subtitle"`
	Orbit     OrbitPlan     `json:"orbit"`
	Frames    []OrbitFrame  `json:"frames"`
	Cues      []CeremonyCue `json:"cues"`
	QuietNote string        `json:"quiet_note"`
}

func NewOrbitPlan(candle Candle, memorial Memorial) OrbitPlan {
	count := 24 + candle.Intensity*4
	if count < 24 {
		count = 24
	}
	if count > 72 {
		count = 72
	}
	radius := 1.0 + float64(candle.Intensity)*0.04
	speed := 0.2 + float64(candle.Intensity)*0.015
	if memorial.Quiet {
		speed *= 0.8
	}
	return OrbitPlan{CandleID: candle.ID, Count: count, Radius: radius, Tilt: float64(len(memorial.Title)%11) * 0.01, Speed: speed}
}

func (p OrbitPlan) Valid() bool {
	return p.CandleID != "" && p.Count >= 12 && p.Count <= 96 && p.Radius > 0 && p.Speed > 0
}

func (p OrbitPlan) Frame(index int) OrbitFrame {
	if p.Count <= 0 {
		return OrbitFrame{}
	}
	index = index % p.Count
	if index < 0 {
		index += p.Count
	}
	progress := float64(index) / float64(p.Count)
	angle := progress*2*math.Pi + p.Tilt
	pulse := 0.5 + 0.5*math.Sin(angle*3+p.Speed)
	return OrbitFrame{Index: index, Angle: angle, Radius: p.Radius + 0.08*math.Sin(angle*2), Brightness: 0.65 + pulse*0.35, Opacity: 0.45 + pulse*0.55}
}

func (p OrbitPlan) Frames(limit int) []OrbitFrame {
	if limit < 1 || limit > p.Count {
		limit = p.Count
	}
	result := make([]OrbitFrame, 0, limit)
	for index := 0; index < limit; index++ {
		result = append(result, p.Frame(index))
	}
	return result
}

func BuildCues(candle Candle, memorial Memorial) []CeremonyCue {
	message := candle.NormalizedMessage()
	if message == "" {
		message = memorial.Dedication
	}
	message = strings.TrimSpace(message)
	first := message
	if len([]rune(first)) > 32 {
		first = string([]rune(first)[:32])
	}
	return []CeremonyCue{
		{At: 0, Kind: "arrival", Message: "愿你在烛光中停留片刻", Duration: 2},
		{At: 2, Kind: "light", Message: first, Duration: 6},
		{At: 8, Kind: "remember", Message: memorial.Dedication, Duration: 10},
		{At: 18, Kind: "return", Message: "星点归于安静，记忆仍在", Duration: 4},
	}
}

func BuildScenePlan(m Memorial, c Candle) ScenePlan {
	plan := NewOrbitPlan(c, m)
	title := strings.TrimSpace(m.Title)
	if title == "" {
		title = "纪念馆"
	}
	subtitle := strings.TrimSpace(c.Message)
	if subtitle == "" {
		subtitle = m.Dedication
	}
	return ScenePlan{Title: title, Subtitle: subtitle, Orbit: plan, Frames: plan.Frames(minInt(plan.Count, 36)), Cues: BuildCues(c, m), QuietNote: "请让视线慢慢回到烛光中央"}
}

func PlanSummary(plan ScenePlan) string {
	return fmt.Sprintf("%s · %s · %d frames", plan.Title, plan.Orbit.CandleID, len(plan.Frames))
}

func SortCues(cues []CeremonyCue) []CeremonyCue {
	ordered := append([]CeremonyCue(nil), cues...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].At == ordered[j].At {
			return ordered[i].Kind < ordered[j].Kind
		}
		return ordered[i].At < ordered[j].At
	})
	return ordered
}

func CueAt(cues []CeremonyCue, tick int) (CeremonyCue, bool) {
	for _, cue := range cues {
		if cue.At == tick {
			return cue, true
		}
	}
	return CeremonyCue{}, false
}

func InterpolateFrame(left, right OrbitFrame, progress float64) OrbitFrame {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return OrbitFrame{Index: left.Index, Angle: left.Angle + (right.Angle-left.Angle)*progress, Radius: left.Radius + (right.Radius-left.Radius)*progress, Brightness: left.Brightness + (right.Brightness-left.Brightness)*progress, Opacity: left.Opacity + (right.Opacity-left.Opacity)*progress}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func ClampFloat(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func BlendColor(primary, secondary string, weight float64) string {
	weight = ClampFloat(weight, 0, 1)
	if weight < 0.5 {
		return primary
	}
	return secondary
}
