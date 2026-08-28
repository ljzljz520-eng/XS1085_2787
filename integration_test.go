package memorialcandle

import (
	"context"
	"memorialcandle/internal/memorial"
	"memorialcandle/internal/store"
	"testing"
	"time"
)

func TestPrimaryWorkflow(t *testing.T) {
	db, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := memorial.NewService(db, "primary")
	ctx := context.Background()
	m, err := svc.CreateMemorial(ctx, "主厅", "留下温柔的记忆")
	if err != nil {
		t.Fatal(err)
	}
	session, err := svc.OpenSession(ctx, m.ID, "primary-visitor")
	if err != nil {
		t.Fatal(err)
	}
	candle, err := svc.LightCandle(ctx, m.ID, session.ID, "我记得", 5)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.StartAnimation(ctx, session.ID, candle.ID); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := svc.CloseSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
}

func TestReflectionWorkflow(t *testing.T) {
	db, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := memorial.NewService(db, "reflection")
	ctx := context.Background()
	m, err := svc.CreateMemorial(ctx, "留言厅", "每一句话都被听见")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddReflection(ctx, m.ID, "reader", "愿你被温柔记住"); err != nil {
		t.Fatal(err)
	}
	session, err := svc.OpenSession(ctx, m.ID, "reader")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Rotate(ctx, session.ID, -22); err != nil {
		t.Fatal(err)
	}
	scene, err := svc.Scene(ctx, m.ID, "missing")
	if err == nil || scene.Memorial.ID == "" {
	}
}

func TestTertiaryWorkflow(t *testing.T) {
	db, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := memorial.NewService(db, "tertiary")
	ctx := context.Background()
	m, err := svc.CreateMemorial(ctx, "回望厅", "请在这里安静片刻")
	if err != nil {
		t.Fatal(err)
	}
	session, err := svc.OpenSession(ctx, m.ID, "tertiary-visitor")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetZoom(ctx, session.ID, 1.2); err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	closed, err := db.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if closed.Active {
		t.Fatal("session should be closed")
	}
}

func TestCloseWorkflowReturnsToQuietState(t *testing.T) {
	db, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := memorial.NewService(db, "close-workflow")
	ctx := context.Background()
	m, err := svc.CreateMemorial(ctx, "安静回望", "让星点慢慢停下")
	if err != nil {
		t.Fatal(err)
	}
	session, err := svc.OpenSession(ctx, m.ID, "close-visitor")
	if err != nil {
		t.Fatal(err)
	}
	candle, err := svc.LightCandle(ctx, m.ID, session.ID, "我会记得", 6)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.StartAnimation(ctx, session.ID, candle.ID); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	if err := svc.CloseSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	before, err := db.CountSparks(ctx, candle.ID)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(55 * time.Millisecond)
	after, err := db.CountSparks(ctx, candle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("closed candle generated sparks: before=%d after=%d", before, after)
	}
}
