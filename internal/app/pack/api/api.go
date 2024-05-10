package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/internal/pkg/health"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	packpb "github.com/dennishilgert/apollo/internal/pkg/proto/pack/v1"
	"github.com/dennishilgert/apollo/internal/pkg/storage"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("apollo.api")

type Options struct {
	Port       int
	WorkerUuid string
}

type Server interface {
	Run(ctx context.Context, healthStatusProvider health.Provider) error
	Ready(ctx context.Context) error
}

type apiServer struct {
	packpb.UnimplementedPackageServiceServer

	port           int
	readyCh        chan struct{}
	running        atomic.Bool
	storageService storage.StorageService
}

// NewApiServer creates a new Server.
func NewApiServer(storageService storage.StorageService, opts Options) Server {
	return &apiServer{
		port:           opts.Port,
		readyCh:        make(chan struct{}),
		storageService: storageService,
	}
}

// Run runs the api server.
func (a *apiServer) Run(ctx context.Context, healthStatusProvider health.Provider) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("api server is already running")
	}

	log.Infof("starting api server on port %d", a.port)
	server := grpc.NewServer()
	packpb.RegisterPackageServiceServer(server, a)

	healthServer := health.NewHealthServer(healthStatusProvider, log)
	healthServer.Register(server)

	lis, lErr := net.Listen("tcp", ":"+fmt.Sprint(a.port))
	if lErr != nil {
		log.Error("error while starting tcp listener")
		return lErr
	}
	// Close the ready channel to signalize that the api server is ready.
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // Ensure channel is closed to avoid goroutine leak.

		if err := server.Serve(lis); err != nil {
			log.Error("error while serving api server")
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// Block until the context is done or an error occurs.
	var serveErr error
	select {
	case <-ctx.Done():
		log.Info("shutting down api server")
	case err := <-errCh: // Handle errors that might have occurred during Serve.
		if err != nil {
			serveErr = err
			log.Errorf("error while listening for requests")
		}
	}

	// Perform graceful shutdown and close the listener regardless of the select outcome.
	stopped := make(chan bool, 1)
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(5 * time.Second)
	select {
	case <-timer.C:
		log.Warn("api server did not stop gracefully in time - forcing shutdown")
		server.Stop()
	case <-stopped:
		timer.Stop()
	}

	if cErr := lis.Close(); cErr != nil && !errors.Is(cErr, net.ErrClosed) && serveErr == nil {
		log.Error("error while closing api server listener")
		return cErr
	}

	return serveErr
}

// Ready waits until the api server is ready or the context is cancelled due to timeout.
func (a *apiServer) Ready(ctx context.Context) error {
	select {
	case <-a.readyCh:
		return nil
	case <-ctx.Done():
		return errors.New("timeout while waiting for the api server to be ready")
	}
}

func (a *apiServer) PresignedUploadUrl(ctx context.Context, req *packpb.PresignedUploadUrlRequest) (*packpb.PresignedUploadUrlResponse, error) {
	url, err := a.storageService.PresignUpload(ctx, req.BucketName, req.ObjectName, time.Duration(req.ExpiresIn)*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload url: %w", err)
	}
	return &packpb.PresignedUploadUrlResponse{
		RawUrl: url.String(),
	}, nil
}
