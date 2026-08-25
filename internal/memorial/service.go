package memorial

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNotFound = errors.New("memorial not found")
	ErrInvalid  = errors.New("invalid memorial input")
	ErrClosed   = errors.New("session already closed")
)

type Service struct {
	repo       Repository
	ids        IDSource
	mu         sync.RWMutex
	animations map[string]*SparkAnimation
	listeners  map[string][]chan SceneEvent
}

func NewService(repo Repository, seed string) *Service {
	return &Service{repo: repo, ids: NewIDSource(seed), animations: map[string]*SparkAnimation{}, listeners: map[string][]chan SceneEvent{}}
}

func (s *Service) CreateMemorial(ctx context.Context, title, dedication string) (Memorial, error) {
	if len(title) < 2 || len(title) > 120 || len(dedication) < 2 || len(dedication) > 600 {
		return Memorial{}, ErrInvalid
	}
	m := Memorial{ID: s.ids.ID("memorial", 1), Title: title, Dedication: dedication, CreatedAt: fixedNow(), Quiet: true}
	if err := s.repo.SaveMemorial(ctx, m); err != nil {
		return Memorial{}, err
	}
	_ = s.repo.SaveAudit(ctx, AuditEntry{ID: s.ids.ID("audit", 1), MemorialID: m.ID, Action: "create", Subject: m.ID, Detail: "memorial created", CreatedAt: fixedNow()})
	return m, nil
}

func (s *Service) OpenSession(ctx context.Context, memorialID, visitorID string) (VisitorSession, error) {
	if memorialID == "" || visitorID == "" {
		return VisitorSession{}, ErrInvalid
	}
	if _, err := s.repo.GetMemorial(ctx, memorialID); err != nil {
		return VisitorSession{}, err
	}
	session := VisitorSession{ID: visitorID, MemorialID: memorialID, StartedAt: fixedNow(), LastSeenAt: fixedNow(), Angle: 0, Zoom: 1, Active: true}
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return VisitorSession{}, err
	}
	s.mu.Lock()
	s.listeners[session.ID] = nil
	s.mu.Unlock()
	return session, nil
}

func (s *Service) Rotate(ctx context.Context, sessionID string, delta float64) (VisitorSession, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return VisitorSession{}, err
	}
	if !session.Active {
		return VisitorSession{}, ErrClosed
	}
	session.Angle = normalizeAngle(session.Angle + delta)
	session.LastSeenAt = fixedNow()
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return VisitorSession{}, err
	}
	s.publish(sessionID, SceneEvent{Kind: "rotate", SessionID: sessionID})
	return session, nil
}

func (s *Service) SetZoom(ctx context.Context, sessionID string, zoom float64) (VisitorSession, error) {
	if zoom < 0.75 || zoom > 3 {
		return VisitorSession{}, ErrInvalid
	}
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return VisitorSession{}, err
	}
	if !session.Active {
		return VisitorSession{}, ErrClosed
	}
	session.Zoom = zoom
	session.LastSeenAt = fixedNow()
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return VisitorSession{}, err
	}
	return session, nil
}

func (s *Service) LightCandle(ctx context.Context, memorialID, visitorID, message string, intensity int) (Candle, error) {
	if len(message) > 240 || intensity < 1 || intensity > 10 {
		return Candle{}, ErrInvalid
	}
	if _, err := s.repo.GetMemorial(ctx, memorialID); err != nil {
		return Candle{}, err
	}
	c := Candle{ID: s.ids.ID("candle", len(message)+intensity), MemorialID: memorialID, VisitorID: visitorID, Message: message, Color: colorForIntensity(intensity), Intensity: intensity, LitAt: fixedNow()}
	if err := s.repo.SaveCandle(ctx, c); err != nil {
		return Candle{}, err
	}
	_ = s.repo.SaveAudit(ctx, AuditEntry{ID: s.ids.ID("audit", intensity+2), MemorialID: memorialID, Action: "light", Subject: c.ID, Detail: "candle lit", CreatedAt: fixedNow()})
	return c, nil
}

func (s *Service) StartAnimation(ctx context.Context, sessionID, candleID string) error {
	if _, err := s.repo.GetSession(ctx, sessionID); err != nil {
		return err
	}
	if _, err := s.repo.GetCandle(ctx, candleID); err != nil {
		return err
	}
	s.mu.Lock()
	if old := s.animations[sessionID]; old != nil {
		old.Stop()
	}
	animation := NewSparkAnimation(ctx, candleID, func(spark Spark) {
		_ = s.repo.SaveSpark(context.Background(), spark)
		s.publish(sessionID, SceneEvent{Kind: "spark", SessionID: sessionID, CandleID: candleID, Spark: &spark})
	})
	s.animations[sessionID] = animation
	s.mu.Unlock()
	animation.Start()
	return nil
}

func (s *Service) CloseSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if !session.Active {
		return nil
	}
	session.Active = false
	session.LastSeenAt = fixedNow()
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return err
	}
	s.mu.Lock()
	animation := s.animations[sessionID]
	if animation != nil {
		animation.Stop()
	}
	delete(s.animations, sessionID)
	s.mu.Unlock()
	s.publish(sessionID, SceneEvent{Kind: "quiet", SessionID: sessionID, Quiet: true})
	return s.repo.SaveAudit(ctx, AuditEntry{ID: s.ids.ID("audit", 99), MemorialID: session.MemorialID, Action: "close", Subject: sessionID, Detail: "visitor returned to quiet state", CreatedAt: fixedNow()})
}

func (s *Service) AddReflection(ctx context.Context, memorialID, visitorID, text string) (Reflection, error) {
	r := Reflection{ID: fmt.Sprintf("reflection-%d", len(text)), MemorialID: memorialID, VisitorID: visitorID, Text: text, Moderated: false, CreatedAt: fixedNow()}
	if !r.Valid() {
		return Reflection{}, ErrInvalid
	}
	if err := s.repo.SaveReflection(ctx, r); err != nil {
		return Reflection{}, err
	}
	return r, nil
}

func (s *Service) Subscribe(sessionID string) <-chan SceneEvent {
	ch := make(chan SceneEvent, 16)
	s.mu.Lock()
	s.listeners[sessionID] = append(s.listeners[sessionID], ch)
	s.mu.Unlock()
	return ch
}

func (s *Service) publish(sessionID string, event SceneEvent) {
	s.mu.RLock()
	listeners := append([]chan SceneEvent(nil), s.listeners[sessionID]...)
	s.mu.RUnlock()
	for _, ch := range listeners {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *Service) Scene(ctx context.Context, memorialID, candleID string) (Scene, error) {
	m, err := s.repo.GetMemorial(ctx, memorialID)
	if err != nil {
		return Scene{}, err
	}
	c, err := s.repo.GetCandle(ctx, candleID)
	if err != nil {
		return Scene{}, err
	}
	sparks, err := s.repo.ListSparks(ctx, candleID, 24)
	if err != nil {
		return Scene{}, err
	}
	reflections, err := s.repo.ListReflections(ctx, memorialID)
	if err != nil {
		return Scene{}, err
	}
	visitors, err := s.repo.CountVisitors(ctx, memorialID)
	if err != nil {
		return Scene{}, err
	}
	return Scene{Memorial: m, Candle: c, Sparks: sparks, Visitors: visitors, Reflections: reflections}, nil
}

func normalizeAngle(value float64) float64 {
	for value >= 360 {
		value -= 360
	}
	for value < 0 {
		value += 360
	}
	return value
}

func colorForIntensity(intensity int) string {
	switch {
	case intensity >= 8:
		return "amber"
	case intensity >= 4:
		return "gold"
	default:
		return "ivory"
	}
}
