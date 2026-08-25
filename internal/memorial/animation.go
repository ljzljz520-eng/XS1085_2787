package memorial

import (
	"context"
	"math"
	"sync"
	"time"
)

type SparkAnimation struct {
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	mu       sync.RWMutex
	closed   bool
	sequence int
	candleID string
	emitter  func(Spark)
}

func NewSparkAnimation(parent context.Context, candleID string, emitter func(Spark)) *SparkAnimation {
	ctx, cancel := context.WithCancel(parent)
	return &SparkAnimation{ctx: ctx, cancel: cancel, done: make(chan struct{}), candleID: candleID, emitter: emitter}
}

func (a *SparkAnimation) Start() {
	go a.run()
}

func (a *SparkAnimation) run() {
	ticker := time.NewTicker(12 * time.Millisecond)
	defer ticker.Stop()
	defer close(a.done)
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.mu.Lock()
			a.sequence++
			seq := a.sequence
			a.mu.Unlock()
			phase := float64(seq%120) / 120 * math.Pi * 2
			spark := Spark{ID: a.candleID + "-spark-" + stringID(seq), CandleID: a.candleID, Sequence: seq, Angle: phase, Radius: 1 + math.Sin(phase)*0.2, Brightness: 0.7 + math.Cos(phase)*0.3, CreatedAt: fixedNow()}
			if a.emitter != nil {
				a.emitter(spark)
			}
		}
	}
}

func (a *SparkAnimation) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	// The closed flag doubles as the guard that makes cancel idempotent:
	// the run loop only exits via the cancelled context, so unless we
	// cancel here the goroutine keeps emitting sparks forever.
	if a.closed {
		return
	}
	a.closed = true
	a.cancel()
}

func (a *SparkAnimation) Wait() {
	<-a.done
}

func (a *SparkAnimation) Closed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.closed
}

func (a *SparkAnimation) Sequence() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sequence
}

func stringID(n int) string {
	if n < 10 {
		return "00" + strconvSmall(n)
	}
	if n < 100 {
		return "0" + strconvSmall(n)
	}
	return strconvSmall(n)
}

func strconvSmall(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	buf := [12]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	return string(buf[i:])
}
