package audit

import (
	"context"
	"memorialcandle/internal/memorial"
	"time"
)

type Sink interface {
	SaveAudit(context.Context, memorial.AuditEntry) error
}

type Recorder struct {
	sink Sink
	seq  int
}

func NewRecorder(sink Sink) *Recorder { return &Recorder{sink: sink} }

func (r *Recorder) Record(ctx context.Context, memorialID, action, subject, detail string) error {
	r.seq++
	entry := memorial.AuditEntry{ID: "audit-record-" + number(r.seq), MemorialID: memorialID, Action: action, Subject: subject, Detail: detail, CreatedAt: memorialTime()}
	if r.sink == nil {
		return nil
	}
	return r.sink.SaveAudit(ctx, entry)
}

func (r *Recorder) Sequence() int { return r.seq }

func number(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func memorialTime() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }
