package memorial

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCandleStopsAfterClose(t *testing.T) {
	var count atomic.Int32
	animation := NewSparkAnimation(context.Background(), "candle-test", func(Spark) { count.Add(1) })
	animation.Start()
	time.Sleep(45 * time.Millisecond)
	animation.Stop()
	before := count.Load()
	time.Sleep(90 * time.Millisecond)
	if count.Load() != before {
		t.Fatalf("spark stream continued after close: before=%d after=%d", before, count.Load())
	}
}
