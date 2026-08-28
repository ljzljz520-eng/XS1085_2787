package ritual

type Viewport struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Angle  float64 `json:"angle"`
	Zoom   float64 `json:"zoom"`
}

func NewViewport(width, height int) Viewport {
	if width < 320 {
		width = 320
	}
	if height < 240 {
		height = 240
	}
	return Viewport{Width: width, Height: height, Zoom: 1}
}

func (v Viewport) Drag(dx, dy float64) Viewport {
	v.Angle += dx*0.15 - dy*0.02
	for v.Angle >= 360 {
		v.Angle -= 360
	}
	for v.Angle < 0 {
		v.Angle += 360
	}
	return v
}

func (v Viewport) Scale(delta float64) Viewport {
	v.Zoom += delta
	if v.Zoom < 0.75 {
		v.Zoom = 0.75
	}
	if v.Zoom > 3 {
		v.Zoom = 3
	}
	return v
}

func (v Viewport) Center() (float64, float64) {
	return float64(v.Width) / 2, float64(v.Height) / 2
}
