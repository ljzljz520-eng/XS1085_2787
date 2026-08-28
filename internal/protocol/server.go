package protocol

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"memorialcandle/internal/memorial"
)

type Server struct {
	service *memorial.Service
	logger  *log.Logger
}

func NewServer(service *memorial.Service, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.Default()
	}
	return &Server{service: service, logger: logger}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/memorials", s.memorials)
	mux.HandleFunc("/sessions", s.sessions)
	mux.HandleFunc("/candles", s.candles)
	mux.HandleFunc("/reflections", s.reflections)
	mux.HandleFunc("/scene", s.scene)
	mux.HandleFunc("/animation/start", s.startAnimation)
	mux.HandleFunc("/animation/close", s.closeAnimation)
	mux.HandleFunc("/sessions/rotate", s.rotate)
	mux.HandleFunc("/sessions/zoom", s.zoom)
	s.extendedRoutes(mux)
	return accessLog(enforceJSON(mux), s.logger)
}

func accessLog(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		logger.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "quiet-ready"})
}

func decode(r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func pathID(path, prefix string) string { return strings.Trim(strings.TrimPrefix(path, prefix), "/") }
