package subpub

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishSubscribe_OrderAndFanout(t *testing.T) {
	sp := NewSubPubWithBuffer(8)

	const n = 100
	wg := sync.WaitGroup{}
	wg.Add(2)

	recv1, recv2 := make([]int, 0, n), make([]int, 0, n)

	_, _ = sp.Subscribe("topic", func(m interface{}) {
		recv1 = append(recv1, m.(int))
		if len(recv1) == n {
			wg.Done()
		}
	})
	_, _ = sp.Subscribe("topic", func(m interface{}) {
		recv2 = append(recv2, m.(int))
		if len(recv2) == n {
			wg.Done()
		}
	})

	for i := 0; i < n; i++ {
		_ = sp.Publish("topic", i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		if recv1[i] != i || recv2[i] != i {
			t.Fatalf("order broken at %d => %d / %d", i, recv1[i], recv2[i])
		}
	}
}

func TestSlowSubscriberDoesNotAffectFast(t *testing.T) {
	sp := NewSubPubWithBuffer(1)

	fast := make(chan int, 10)

	_, _ = sp.Subscribe("slow", func(m interface{}) { time.Sleep(20 * time.Millisecond) })

	_, _ = sp.Subscribe("slow", func(m interface{}) { fast <- 1 })

	start := time.Now()
	_ = sp.Publish("slow", 42)
	_ = sp.Publish("slow", 43)

	select {
	case <-fast:
	case <-time.After(10 * time.Millisecond):
		t.Fatal("fast subscriber blocked by slow ")
	}
	elapsed := time.Since(start)
	if elapsed > 40*time.Millisecond {
		t.Fatalf("publish blocked for %v", elapsed)
	}
}

func TestUnsubscribe(t *testing.T) {
	sp := NewSubPub()

	var cnt atomic.Int32
	sub, _ := sp.Subscribe("x", func(m interface{}) {
		cnt.Add(1)
	})

	_ = sp.Publish("x", 1)
	time.Sleep(time.Millisecond)

	sub.Unsubscribe()
	_ = sp.Publish("x", 2)
	time.Sleep(time.Millisecond)

	if cnt.Load() != 1 {
		t.Fatalf("expected 1 msg, got %d", cnt.Load())
	}
}

func TestCloseWithContext(t *testing.T) {
	sp := NewSubPub()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := sp.Close(ctx); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if err := sp.Close(context.Background()); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}
