package store

import (
	"context"
	"memorialcandle/internal/memorial"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memorial.db")
	ctx := context.Background()
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	value := memorial.Memorial{ID: "memorial-reopen", Title: "重开的烛光", Dedication: "数据仍然在这里", CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), Quiet: true}
	if err := first.SaveMemorial(ctx, value); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetMemorial(ctx, value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != value.Title || got.Dedication != value.Dedication {
		t.Fatalf("reopened value mismatch: %#v", got)
	}
}
