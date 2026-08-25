package memorial

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

type JobRepository interface {
	SaveAnimationJob(context.Context, AnimationJob) error
	GetAnimationJob(context.Context, string) (AnimationJob, error)
}

type AuditReader interface {
	ListAuditEntries(context.Context, string, int) ([]AuditEntry, error)
}

type Timeline struct {
	SessionID string       `json:"session_id"`
	CandleID  string       `json:"candle_id"`
	State     CandleState  `json:"state"`
	Job       AnimationJob `json:"job"`
	Events    []SceneEvent `json:"events"`
	StartedAt time.Time    `json:"started_at"`
	ClosedAt  time.Time    `json:"closed_at"`
}

type ServiceOptions struct {
	AnimationLimit int
	QuietDelay     time.Duration
	Clock          func() time.Time
}

func DefaultServiceOptions() ServiceOptions {
	return ServiceOptions{AnimationLimit: 120, QuietDelay: 12 * time.Millisecond, Clock: fixedNow}
}

func (s *Service) ValidateSession(ctx context.Context, sessionID string) (VisitorSession, error) {
	if strings.TrimSpace(sessionID) == "" {
		return VisitorSession{}, ErrInvalid
	}
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return VisitorSession{}, err
	}
	if !session.Valid() {
		return VisitorSession{}, errors.New("stored session is invalid")
	}
	return session, nil
}

func (s *Service) CandleState(ctx context.Context, candleID string) (CandleState, error) {
	candle, err := s.repo.GetCandle(ctx, candleID)
	if err != nil {
		return "", err
	}
	if candle.Extinguished {
		return CandleClosed, nil
	}
	s.mu.RLock()
	for _, animation := range s.animations {
		if animation != nil && animation.candleID == candleID {
			s.mu.RUnlock()
			return CandleSpreading, nil
		}
	}
	s.mu.RUnlock()
	return CandleQuiet, nil
}

func (s *Service) BuildPlan(ctx context.Context, memorialID, candleID string) (ScenePlan, error) {
	memorialValue, err := s.repo.GetMemorial(ctx, memorialID)
	if err != nil {
		return ScenePlan{}, err
	}
	candle, err := s.repo.GetCandle(ctx, candleID)
	if err != nil {
		return ScenePlan{}, err
	}
	if candle.MemorialID != memorialValue.ID {
		return ScenePlan{}, errors.New("candle belongs to another memorial")
	}
	return BuildScenePlan(memorialValue, candle), nil
}

func (s *Service) StartTrackedAnimation(ctx context.Context, sessionID, candleID string) (AnimationJob, error) {
	if err := s.StartAnimation(ctx, sessionID, candleID); err != nil {
		return AnimationJob{}, err
	}
	job := NewAnimationJob(s.ids.ID("animation", len(sessionID)+len(candleID)), candleID, fixedNow())
	if err := job.Start(fixedNow()); err != nil {
		return AnimationJob{}, err
	}
	if repository, ok := s.repo.(JobRepository); ok {
		if err := repository.SaveAnimationJob(ctx, job); err != nil {
			return AnimationJob{}, err
		}
	}
	return job, nil
}

func (s *Service) StopTrackedAnimation(ctx context.Context, sessionID string) (AnimationJob, error) {
	session, err := s.ValidateSession(ctx, sessionID)
	if err != nil && !errors.Is(err, ErrClosed) {
		return AnimationJob{}, err
	}
	s.mu.Lock()
	animation := s.animations[sessionID]
	if animation != nil {
		animation.Stop()
	}
	delete(s.animations, sessionID)
	s.mu.Unlock()
	job := AnimationJob{ID: s.ids.ID("animation", len(sessionID)), CandleID: session.MemorialID, Status: AnimationStopped, Generated: 0, StartedAt: fixedNow(), StoppedAt: fixedNow()}
	if repository, ok := s.repo.(JobRepository); ok {
		if err := repository.SaveAnimationJob(ctx, job); err != nil {
			return AnimationJob{}, err
		}
	}
	return job, nil
}

func (s *Service) Timeline(ctx context.Context, sessionID, candleID string) (Timeline, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return Timeline{}, err
	}
	candle, err := s.repo.GetCandle(ctx, candleID)
	if err != nil {
		return Timeline{}, err
	}
	if candle.MemorialID != session.MemorialID {
		return Timeline{}, errors.New("session and candle are not paired")
	}
	state, err := s.CandleState(ctx, candleID)
	if err != nil {
		return Timeline{}, err
	}
	job := AnimationJob{ID: s.ids.ID("animation", len(sessionID)+len(candleID)), CandleID: candleID, Status: AnimationPending, StartedAt: fixedNow()}
	if repository, ok := s.repo.(JobRepository); ok {
		if stored, storedErr := repository.GetAnimationJob(ctx, job.ID); storedErr == nil {
			job = stored
		}
	}
	return Timeline{SessionID: sessionID, CandleID: candleID, State: state, Job: job, StartedAt: session.StartedAt}, nil
}

func (s *Service) RecentEvents(ctx context.Context, memorialID string, limit int) ([]SceneEvent, error) {
	if limit <= 0 {
		limit = 16
	}
	if limit > 128 {
		limit = 128
	}
	if reader, ok := s.repo.(AuditReader); ok {
		entries, err := reader.ListAuditEntries(ctx, memorialID, limit)
		if err != nil {
			return nil, err
		}
		events := make([]SceneEvent, 0, len(entries))
		for _, entry := range entries {
			events = append(events, SceneEvent{Kind: entry.Action, CandleID: entry.Subject, Quiet: entry.Action == "close"})
		}
		return events, nil
	}
	return []SceneEvent{}, nil
}

func SortSceneEvents(events []SceneEvent) []SceneEvent {
	result := append([]SceneEvent(nil), events...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Kind == result[j].Kind {
			return result[i].CandleID < result[j].CandleID
		}
		return result[i].Kind < result[j].Kind
	})
	return result
}

func MergeEvents(primary, secondary []SceneEvent) []SceneEvent {
	result := make([]SceneEvent, 0, len(primary)+len(secondary))
	result = append(result, primary...)
	result = append(result, secondary...)
	return SortSceneEvents(result)
}

func (s *Service) CloseAndWait(ctx context.Context, sessionID string) error {
	// Capture the running animation before CloseSession deletes it from the
	// map. Otherwise the lookup below reads nil and we return before the
	// goroutine has actually stopped — letting sparks keep being generated.
	s.mu.RLock()
	animation := s.animations[sessionID]
	s.mu.RUnlock()
	if err := s.CloseSession(ctx, sessionID); err != nil {
		return err
	}
	if animation != nil {
		animation.Wait()
	}
	return nil
}

func (s *Service) LockSnapshot() func() {
	s.mu.RLock()
	return s.mu.RUnlock
}
