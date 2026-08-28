package memorial_test

import (
	"context"
	"memorialcandle/internal/memorial"
	"memorialcandle/internal/store"
	"testing"
	"time"
)

func TestServiceLightsAndRotates(t *testing.T) {
	db, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := memorial.NewService(db, "service-test")
	ctx := context.Background()
	m, err := svc.CreateMemorial(ctx, "冬日纪念馆", "愿记忆保留温度")
	if err != nil {
		t.Fatal(err)
	}
	session, err := svc.OpenSession(ctx, m.ID, "visitor-one")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Rotate(ctx, session.ID, 42); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetZoom(ctx, session.ID, 1.4); err != nil {
		t.Fatal(err)
	}
	candle, err := svc.LightCandle(ctx, m.ID, session.ID, "谢谢你来过", 7)
	if err != nil {
		t.Fatal(err)
	}
	if candle.Color != "gold" {
		t.Fatalf("unexpected color %q", candle.Color)
	}
	if err := svc.StartAnimation(ctx, session.ID, candle.ID); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	if err := svc.CloseSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
}
