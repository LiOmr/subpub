package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/li-omr/subpub/internal/config"
	gHandler "github.com/li-omr/subpub/internal/handler/grpc"
	"github.com/li-omr/subpub/internal/logger"
	"github.com/li-omr/subpub/internal/service"
	pb "github.com/li-omr/subpub/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.Log.Level)
	defer log.Sync()

	svc := service.New(cfg.SubPub.Buffer)

	gsrv := grpc.NewServer()
	pb.RegisterPubSubServer(gsrv, gHandler.New(svc, log))
	reflection.Register(gsrv)

	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen", zap.Error(err))
	}
	go func() {
		log.Info("gRPC serving", zap.String("addr", addr))
		if err := gsrv.Serve(lis); err != nil {
			log.Fatal("serve", zap.Error(err))
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutting down")
	gsrv.GracefulStop()
	_ = svc.Close(context.Background())
}
