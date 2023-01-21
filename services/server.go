package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"paydex/assets"
	"paydex/config"
	pb "paydex/pkg/gen"
	"paydex/worker"
	"time"

	_ "expvar"         // Register the expvar handlers
	_ "net/http/pprof" // Register the pprof handlers

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedPaydexServiceServer
	worker   worker.TaskDistributor
	cfg      *config.Config
	redisOpt asynq.RedisClientOpt
	l        *slog.Logger
}

func NewServer(
	worker worker.TaskDistributor,
	cfg *config.Config,
	l *slog.Logger,
	redisOpt asynq.RedisClientOpt) *Server {
	return &Server{
		worker:   worker,
		cfg:      cfg,
		l:        l,
		redisOpt: redisOpt,
	}
}

func (s *Server) RunGrpcServer() error {
	dsn := fmt.Sprintf("%s:%s", s.cfg.Servers["grpc"].Address, s.cfg.Servers["grpc"].Port)

	lis, err := net.Listen("tcp", dsn)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			otelgrpc.StreamServerInterceptor(),
			// grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			otelgrpc.UnaryServerInterceptor(),
		)))

	pb.RegisterPaydexServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Print("grpc sever started")

	return grpcServer.Serve(lis)
}

func (s *Server) RunHTTPServer() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dsn := fmt.Sprintf("%s:%s", s.cfg.Servers["http"].Address, s.cfg.Servers["http"].Port)
	// dial the gRPC server above to make a client connection
	conn, err := grpc.Dial(dsn, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("fail to dial: %w", err)
	}
	defer conn.Close()

	// create an HTTP router using the client connection above
	// and register it with the service client
	rmux := runtime.NewServeMux()
	client := pb.NewPaydexServiceClient(conn)
	err = pb.RegisterPaydexServiceHandlerClient(ctx, rmux, client)
	if err != nil {
		return err
	}

	// create a standard HTTP router
	mux := http.NewServeMux()
	// mount the gRPC HTTP gateway to the root
	mux.Handle("/", rmux)

	// mount a path to expose the generated OpenAPI specification on disk
	mux.HandleFunc("/swagger-ui/paydex.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./gen/protos/service.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.FS(assets.EmbeddedFiles))))

	srv := &http.Server{
		Addr:              "localhost:8080",
		Handler:           mux,
		TLSConfig:         nil,
		ReadTimeout:       time.Duration(s.cfg.Servers["http"].Timeout) * time.Second,
		ReadHeaderTimeout: time.Duration(s.cfg.Servers["http"].Timeout) * time.Second,
		WriteTimeout:      time.Duration(s.cfg.Servers["http"].Timeout) * time.Second,
		IdleTimeout:       time.Duration(s.cfg.Servers["http"].Timeout) * time.Second,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
	log.Print("http sever started")

	//Register debug handlers
	if !s.cfg.Prod {
		log.Printf("system is in debug mode: running debug servers @http://localhost:8091")
		debugServer := NewDebugServer("localhost:8091")
		go func() {
			log.Fatal(debugServer.ListenAndServe())
		}()
	}

	// start a standard HTTP server with the router
	return srv.ListenAndServe()
}

func (s *Server) RunTaskProcessor() error {
	taskProcessor := worker.NewRedisTaskProcessor(s.redisOpt, s.cfg)
	slog.Info("start task processor")
	if err := taskProcessor.Start(); err != nil {
		slog.Error("failed to start task processor", err)
		return err
	}
	if err := taskProcessor.StartScheduler(); err != nil {
		slog.Error("failed to start task scheduler", err)
		return err
	}
	return nil
}

type DebugServer struct {
	*http.Server
}

// NewDebugServer provides new debug http server
func NewDebugServer(address string) *DebugServer {
	return &DebugServer{
		&http.Server{
			Addr:        address,
			Handler:     http.DefaultServeMux,
			ReadTimeout: 1 * time.Second,
		},
	}
}
