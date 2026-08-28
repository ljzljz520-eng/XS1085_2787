package ritual

import "math"

type Palette struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
	Speed  float64  `json:"speed"`
}

func QuietPalette() Palette {
	return Palette{Name: "quiet-amber", Colors: []string{"#fff7d6", "#f5c46b", "#d98d4c", "#75614c"}, Speed: 0.35}
}

func (p Palette) ColorAt(index int) string {
	if len(p.Colors) == 0 {
		return "#fff7d6"
	}
	return p.Colors[index%len(p.Colors)]
}

func (p Palette) IntensityAt(progress float64) float64 {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return 0.55 + 0.45*math.Sin(progress*math.Pi)
}

func (p Palette) Validate() bool {
	return p.Name != "" && len(p.Colors) >= 3 && p.Speed > 0 && p.Speed < 2
}
