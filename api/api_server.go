package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/model"
	pb "github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/worker"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
)

type server struct {
	pb.UnimplementedStreamServer
}

func (s server) RequestStream(ctx context.Context, request *pb.StreamRequest) (*pb.Status, error) {
	if request.WorkerId != cfg.WorkerID {
		log.Info("Rejected request to stream")
		return &pb.Status{Ok: false}, nil
	}
	return &pb.Status{Ok: true}, nil
}

//InitApi Initializes api endpoints
//addr: port to run on, e.g. ":8080"
func InitApi(addr string) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}
	grpcServer := grpc.NewServer()
	pb.RegisterStreamServer(grpcServer, &server{})

	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// return ok for healthCheck
	router.Handle("GET", "/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	authenticated := router.Group("/worker", validate)

	authenticated.GET("/stream", func(c *gin.Context) {
		var req model.StreamRequest
		err := c.MustBindWith(&req, binding.JSON)
		if err != nil {
			log.WithError(err).Error("Could not bind stream request")
			return
		}
		go worker.HandleStreamRequest(req) // async start streaming
	})

	err = router.Run(addr)
	if err != nil {
		log.WithError(err).Error("Could not initialize api")
		return
	}
}

//validate checks if the HTTP header "X-API-KEY" matches the worker id. Aborts request otherwise.
func validate(c *gin.Context) {
	if os.Getenv("workerID") != c.GetHeader("X-API-Key") {
		log.WithFields(log.Fields{
			"URL": c.Request.RequestURI,
			"IP":  c.ClientIP(),
		}).Info("Rejected request")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
