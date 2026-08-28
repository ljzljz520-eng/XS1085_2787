package protocol

import (
	"encoding/json"
	"errors"
	"memorialcandle/internal/catalog"
	"memorialcandle/internal/insights"
	"memorialcandle/internal/memorial"
	"net/http"
	"strconv"
	"strings"
)

type PlanRequest struct {
	MemorialID string `json:"memorial_id"`
	CandleID   string `json:"candle_id"`
}

type TimelineRequest struct {
	SessionID string `json:"session_id"`
	CandleID  string `json:"candle_id"`
}

type CatalogResponse struct {
	Locale   string            `json:"locale"`
	Messages []catalog.Message `json:"messages"`
}

type EventResponse struct {
	Events []memorial.SceneEvent `json:"events"`
}

type MetricsResponse struct {
	Dashboard insights.Dashboard `json:"dashboard"`
}

func (s *Server) extendedRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/plan", s.plan)
	mux.HandleFunc("/timeline", s.timeline)
	mux.HandleFunc("/events", s.events)
	mux.HandleFunc("/catalog", s.catalog)
	mux.HandleFunc("/metrics", s.metrics)
}

func (s *Server) plan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	memorialID := r.URL.Query().Get("memorial_id")
	candleID := r.URL.Query().Get("candle_id")
	plan, err := s.service.BuildPlan(r.Context(), memorialID, candleID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (s *Server) timeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	timeline, err := s.service.Timeline(r.Context(), r.URL.Query().Get("session_id"), r.URL.Query().Get("candle_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	limit := 16
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	events, err := s.service.RecentEvents(r.Context(), r.URL.Query().Get("memorial_id"), limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, EventResponse{Events: events})
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	value := catalog.New()
	locale := catalog.NormalizeLocale(r.URL.Query().Get("locale"))
	writeJSON(w, http.StatusOK, CatalogResponse{Locale: string(locale), Messages: value.Ordered()})
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	metrics := insights.NewMetrics()
	metrics.SetVisitors(0)
	dashboard := insights.BuildDashboard("纪念馆", metrics, nil)
	writeJSON(w, http.StatusOK, MetricsResponse{Dashboard: dashboard})
}

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, memorial.ErrInvalid) {
		status = http.StatusUnprocessableEntity
	}
	if errors.Is(err, memorial.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, memorial.ErrClosed) {
		status = http.StatusConflict
	}
	writeError(w, status, err.Error())
}

func acceptsJSON(r *http.Request) bool {
	value := r.Header.Get("Accept")
	return value == "" || strings.Contains(value, "application/json") || strings.Contains(value, "*/*")
}

func contentTypeJSON(r *http.Request) bool {
	value := strings.ToLower(r.Header.Get("Content-Type"))
	return value == "" || strings.HasPrefix(value, "application/json")
}

func enforceJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !acceptsJSON(r) {
			writeError(w, http.StatusNotAcceptable, "application/json is required")
			return
		}
		if r.Method == http.MethodPost && !contentTypeJSON(r) {
			writeError(w, http.StatusUnsupportedMediaType, "application/json is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeStrict(r *http.Request, target any) error {
	if !contentTypeJSON(r) {
		return errors.New("content type must be json")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
