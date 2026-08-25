package ritual

import (
	"math"
	"sort"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Polar struct {
	Angle  float64 `json:"angle"`
	Radius float64 `json:"radius"`
}

type Ring struct {
	Center Point   `json:"center"`
	Items  []Point `json:"items"`
}

type Motion struct {
	Start  Point  `json:"start"`
	End    Point  `json:"end"`
	Frames int    `json:"frames"`
	Curve  string `json:"curve"`
}

func NewPoint(x, y, z float64) Point { return Point{X: x, Y: y, Z: z} }

func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y, Z: p.Z + other.Z}
}

func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y, Z: p.Z - other.Z}
}

func (p Point) Scale(factor float64) Point {
	return Point{X: p.X * factor, Y: p.Y * factor, Z: p.Z * factor}
}

func (p Point) Length() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
}

func (p Point) Distance(other Point) float64 { return p.Sub(other).Length() }

func (p Point) Normalize() Point {
	length := p.Length()
	if length == 0 {
		return Point{}
	}
	return p.Scale(1 / length)
}

func PolarPoint(p Polar) Point {
	return Point{X: math.Cos(p.Angle) * p.Radius, Y: math.Sin(p.Angle) * p.Radius}
}

func RingPoints(center Point, count int, radius, phase float64) Ring {
	if count < 1 {
		count = 1
	}
	if radius <= 0 {
		radius = 1
	}
	items := make([]Point, 0, count)
	for index := 0; index < count; index++ {
		angle := phase + float64(index)*2*math.Pi/float64(count)
		point := PolarPoint(Polar{Angle: angle, Radius: radius})
		items = append(items, center.Add(point))
	}
	return Ring{Center: center, Items: items}
}

func LayeredRings(center Point, layers, count int, radius, gap float64) []Ring {
	if layers < 1 {
		layers = 1
	}
	result := make([]Ring, 0, layers)
	for layer := 0; layer < layers; layer++ {
		result = append(result, RingPoints(center, count+layer%3, radius+float64(layer)*gap, float64(layer)*0.12))
	}
	return result
}

func EaseInOut(progress float64) float64 {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return progress * progress * (3 - 2*progress)
}

func EaseSine(progress float64) float64 {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return 0.5 - 0.5*math.Cos(progress*math.Pi)
}

func InterpolatePoint(start, end Point, progress float64) Point {
	progress = EaseInOut(progress)
	return Point{X: start.X + (end.X-start.X)*progress, Y: start.Y + (end.Y-start.Y)*progress, Z: start.Z + (end.Z-start.Z)*progress}
}

func MotionFrames(m Motion) []Point {
	frames := m.Frames
	if frames < 2 {
		frames = 2
	}
	result := make([]Point, 0, frames)
	for index := 0; index < frames; index++ {
		progress := float64(index) / float64(frames-1)
		if m.Curve == "sine" {
			progress = EaseSine(progress)
		}
		result = append(result, InterpolatePoint(m.Start, m.End, progress))
	}
	return result
}

func Rotate(point Point, radians float64) Point {
	cosine := math.Cos(radians)
	sine := math.Sin(radians)
	return Point{X: point.X*cosine - point.Y*sine, Y: point.X*sine + point.Y*cosine, Z: point.Z}
}

func Project(point Point, width, height int, zoom float64) Point {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if zoom <= 0 {
		zoom = 1
	}
	scale := float64(minInt(width, height)) * 0.36 * zoom
	return Point{X: float64(width)/2 + point.X*scale, Y: float64(height)/2 - point.Y*scale, Z: point.Z}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func SortByDepth(points []Point) []Point {
	result := append([]Point(nil), points...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].Z < result[j].Z })
	return result
}

func BoundPoint(point Point, radius float64) Point {
	if radius <= 0 {
		return Point{}
	}
	length := point.Length()
	if length <= radius {
		return point
	}
	return point.Scale(radius / length)
}

func OrbitPoint(index, count int, radius, tilt, phase float64) Point {
	if count < 1 {
		count = 1
	}
	angle := phase + float64(index%count)*2*math.Pi/float64(count)
	point := PolarPoint(Polar{Angle: angle, Radius: radius})
	point.Z = math.Sin(angle+tilt) * radius * 0.18
	return point
}

func OrbitCloud(count int, radius, tilt float64) []Point {
	if count < 1 {
		return []Point{}
	}
	result := make([]Point, 0, count)
	for index := 0; index < count; index++ {
		result = append(result, OrbitPoint(index, count, radius, tilt, 0))
	}
	return result
}
