package protocol

import "memorialcandle/internal/memorial"

type CreateRequest struct {
	Title      string `json:"title"`
	Dedication string `json:"dedication"`
}

type SessionRequest struct {
	MemorialID string `json:"memorial_id"`
	VisitorID  string `json:"visitor_id"`
}

type CandleRequest struct {
	MemorialID string `json:"memorial_id"`
	VisitorID  string `json:"visitor_id"`
	Message    string `json:"message"`
	Intensity  int    `json:"intensity"`
}

type ReflectionRequest struct {
	MemorialID string `json:"memorial_id"`
	VisitorID  string `json:"visitor_id"`
	Text       string `json:"text"`
}

type RotateRequest struct {
	Delta float64 `json:"delta"`
}
type ZoomRequest struct {
	Zoom float64 `json:"zoom"`
}
type AnimationRequest struct {
	SessionID string `json:"session_id"`
	CandleID  string `json:"candle_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type SceneResponse struct {
	Scene memorial.Scene `json:"scene"`
}
