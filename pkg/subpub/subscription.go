package subpub

import (
	"sync"
)

type subscription struct {
	subject string
	ch      chan interface{}
	cb      MessageHandler
	parent  *subPubImpl
	once    sync.Once
}

func (s *subscription) start() {
	s.parent.wg.Add(1)
	go func() {
		defer s.parent.wg.Done()
		for msg := range s.ch {
			s.cb(msg)
		}
	}()
}

func (s *subscription) Unsubscribe() {
	s.once.Do(func() {
		s.parent.mu.Lock()
		delete(s.parent.subs[s.subject], s)
		s.parent.mu.Unlock()
		close(s.ch)
	})
}
