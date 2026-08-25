package audit

import (
	"context"
	"memorialcandle/internal/memorial"
	"testing"
)

type memorySink struct{ entries []memorial.AuditEntry }

func (m *memorySink) SaveAudit(_ context.Context, entry memorial.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func TestRecorderCreatesOrderedEntries(t *testing.T) {
	sink := &memorySink{}
	recorder := NewRecorder(sink)
	if err := recorder.Record(context.Background(), "memorial", "light", "candle", "lit"); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Record(context.Background(), "memorial", "close", "session", "quiet"); err != nil {
		t.Fatal(err)
	}
	if recorder.Sequence() != 2 || len(sink.entries) != 2 {
		t.Fatalf("unexpected audit count %d/%d", recorder.Sequence(), len(sink.entries))
	}
}
