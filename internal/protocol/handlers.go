package protocol

import (
	"memorialcandle/internal/memorial"
	"net/http"
)

func (s *Server) memorials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input CreateRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	value, err := s.service.CreateMemorial(r.Context(), input.Title, input.Dedication)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (s *Server) sessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input SessionRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	value, err := s.service.OpenSession(r.Context(), input.MemorialID, input.VisitorID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (s *Server) candles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input CandleRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	value, err := s.service.LightCandle(r.Context(), input.MemorialID, input.VisitorID, input.Message, input.Intensity)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (s *Server) reflections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input ReflectionRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	value, err := s.service.AddReflection(r.Context(), input.MemorialID, input.VisitorID, input.Text)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (s *Server) scene(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	memorialID := r.URL.Query().Get("memorial_id")
	candleID := r.URL.Query().Get("candle_id")
	value, err := s.service.Scene(r.Context(), memorialID, candleID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, SceneResponse{Scene: value})
}

func (s *Server) startAnimation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input AnimationRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.service.StartAnimation(r.Context(), input.SessionID, input.CandleID); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "spark-expansion"})
}

func (s *Server) closeAnimation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input SessionRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.service.CloseSession(r.Context(), input.VisitorID); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "quiet"})
}

func (s *Server) rotate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input RotateRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	sessionID := r.URL.Query().Get("session_id")
	value, err := s.service.Rotate(r.Context(), sessionID, input.Delta)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (s *Server) zoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var input ZoomRequest
	if err := decode(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	sessionID := r.URL.Query().Get("session_id")
	value, err := s.service.SetZoom(r.Context(), sessionID, input.Zoom)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func statusFor(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if err == memorial.ErrInvalid {
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}
