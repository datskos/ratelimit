package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/datskos/ratelimit/pkg/config"
	"github.com/datskos/ratelimit/pkg/proto"
	"github.com/datskos/ratelimit/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config  config.AppConfig
	storage storage.Storage
	grpc    *grpc.Server
}

func NewServer(config config.AppConfig) (*Server, error) {
	storage, err := storage.NewStorage(config)
	if err != nil {
		return nil, err
	}

	service := NewService(storage)
	grpc := grpc.NewServer()
	proto.RegisterRateLimitServiceServer(grpc, service)
	if config.Reflection {
		reflection.Register(grpc)
	}

	return &Server{
		config:  config,
		storage: storage,
		grpc:    grpc,
	}, nil
}

func (server *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", server.config.Port))
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		err := server.grpc.Serve(lis)
		errCh <- err
	}()
	server.handleShutdown()

	err = <-errCh
	return err
}

func (server *Server) handleShutdown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-ch
		log.Println("received shutdown signal. gracefully shutting down")
		server.grpc.GracefulStop()
		server.storage.Close()
		os.Exit(0)
	}()
}
