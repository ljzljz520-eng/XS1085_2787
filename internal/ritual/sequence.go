package ritual

import "memorialcandle/internal/memorial"

type Ceremony struct {
	Palette Palette  `json:"palette"`
	Steps   []string `json:"steps"`
}

func DefaultCeremony() Ceremony {
	return Ceremony{Palette: QuietPalette(), Steps: []string{"arrive", "light", "remember", "return"}}
}

func (c Ceremony) Step(index int) string {
	if len(c.Steps) == 0 {
		return "arrive"
	}
	if index < 0 {
		index = 0
	}
	return c.Steps[index%len(c.Steps)]
}

func (c Ceremony) Event(sessionID, candleID, kind string) memorial.SceneEvent {
	quiet := kind == "return" || kind == "close"
	return memorial.SceneEvent{Kind: kind, SessionID: sessionID, CandleID: candleID, Quiet: quiet}
}

func (c Ceremony) Valid() bool {
	return c.Palette.Validate() && len(c.Steps) >= 4
}
