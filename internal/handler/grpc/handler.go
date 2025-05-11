package grpc

import (
	"context"

	"github.com/li-omr/subpub/internal/service"
	pb "github.com/li-omr/subpub/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedPubSubServer
	svc *service.Service
	log *zap.Logger
}

func New(svc *service.Service, log *zap.Logger) *Server {
	return &Server{svc: svc, log: log}
}

func (s *Server) Publish(
	ctx context.Context,
	req *pb.PublishRequest,
) (*emptypb.Empty, error) {
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty key")
	}
	if err := s.svc.Publish(req.Key, req.Data); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Subscribe(
	req *pb.SubscribeRequest,
	stream pb.PubSub_SubscribeServer,
) error {
	if req.GetKey() == "" {
		return status.Error(codes.InvalidArgument, "empty key")
	}
	sub, err := s.svc.Subscribe(req.Key, func(d string) {
		_ = stream.Send(&pb.Event{Data: d})
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer sub.Unsubscribe()

	<-stream.Context().Done()
	return stream.Context().Err()
}
