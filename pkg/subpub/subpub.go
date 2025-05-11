package subpub

import (
	"context"
	"sync"
)

const defaultBuf = 1024

func NewSubPub() SubPub { return NewSubPubWithBuffer(defaultBuf) }

func NewSubPubWithBuffer(buffer int) SubPub {
	if buffer <= 0 {
		buffer = defaultBuf
	}
	return &subPubImpl{
		subs:    make(map[string]map[*subscription]struct{}),
		bufSize: buffer,
	}
}

type subPubImpl struct {
	mu      sync.RWMutex
	subs    map[string]map[*subscription]struct{}
	closed  bool
	closeWG sync.Once
	wg      sync.WaitGroup
	bufSize int
}

func (sp *subPubImpl) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.closed {
		return nil, ErrClosed
	}
	sub := &subscription{
		subject: subject,
		ch:      make(chan interface{}, sp.bufSize),
		cb:      cb,
		parent:  sp,
	}
	if _, ok := sp.subs[subject]; !ok {
		sp.subs[subject] = make(map[*subscription]struct{})
	}
	sp.subs[subject][sub] = struct{}{}
	sub.start()
	return sub, nil
}

func (sp *subPubImpl) Publish(subject string, msg interface{}) error {
	sp.mu.RLock()
	if sp.closed {
		sp.mu.RUnlock()
		return ErrClosed
	}

	snapshot := make([]*subscription, 0, len(sp.subs[subject]))
	for s := range sp.subs[subject] {
		snapshot = append(snapshot, s)
	}
	sp.mu.RUnlock()

	for _, sub := range snapshot {
		sub.ch <- msg
	}
	return nil
}

func (sp *subPubImpl) Close(ctx context.Context) error {
	var err error
	sp.closeWG.Do(func() {
		sp.mu.Lock()
		sp.closed = true
		for _, set := range sp.subs {
			for sub := range set {
				close(sub.ch)
			}
		}
		sp.mu.Unlock()

		done := make(chan struct{})
		go func() { sp.wg.Wait(); close(done) }()

		select {
		case <-done:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}
