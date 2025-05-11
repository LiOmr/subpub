package service

import (
	"context"

	"github.com/li-omr/subpub/pkg/subpub"
)

type Service struct{ bus subpub.SubPub }

func New(buf int) *Service { return &Service{bus: subpub.NewSubPubWithBuffer(buf)} }

func (s *Service) Publish(k, d string) error { return s.bus.Publish(k, d) }

func (s *Service) Subscribe(k string, cb func(string)) (subpub.Subscription, error) {
	return s.bus.Subscribe(k, func(m interface{}) {
		if str, ok := m.(string); ok {
			cb(str)
		}
	})
}

func (s *Service) Close(ctx context.Context) error { return s.bus.Close(ctx) }
