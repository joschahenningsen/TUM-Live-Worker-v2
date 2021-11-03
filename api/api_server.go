package api

import (
	"context"
	"errors"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/worker"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type server struct {
	pb.UnimplementedToWorkerServer
}

func (s server) RequestStream(ctx context.Context, request *pb.StreamRequest) (*pb.Status, error) {
	if request.WorkerId != cfg.WorkerID {
		log.Info("Rejected request to stream")
		return &pb.Status{Ok: false}, errors.New("unauthenticated: wrong worker id")
	}
	go worker.HandleStreamRequest(request)
	return &pb.Status{Ok: true}, nil
}

func (s server) RequestPremiere(ctx context.Context, request *pb.PremiereRequest) (*pb.Status, error) {
	if request.WorkerID != cfg.WorkerID {
		log.Info("Rejected request to stream")
		return &pb.Status{Ok: false}, errors.New("unauthenticated: wrong worker id")
	}
	go worker.HandlePremiere(request)
	return &pb.Status{Ok: true}, nil
}

//InitApi Initializes api endpoints
//addr: port to run on, e.g. ":8080"
func InitApi(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}
	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Minute,
		MaxConnectionAge:      time.Minute,
		MaxConnectionAgeGrace: time.Second * 5,
		Time:                  time.Minute * 10,
		Timeout:               time.Second * 20,
	}))
	pb.RegisterToWorkerServer(grpcServer, &server{})

	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
