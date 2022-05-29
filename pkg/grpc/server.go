package grpc

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	grpc_go "google.golang.org/grpc"
	"io"
	"net"
)

// Server is an interface for the dapr gRPC server.
type Server interface {
	io.Closer
	StartNonBlocking() error
}
type server struct {
	api    API
	config ServerConfig

	servers []*grpc_go.Server
}

// NewAPIServer returns a new user facing gRPC API server.
func NewAPIServer(api API, config ServerConfig) Server {
	return &server{
		api:    api,
		config: config,
	}
}

// StartNonBlocking starts a new server in a goroutine.
func (s *server) StartNonBlocking() error {
	var listeners []net.Listener
	for _, apiListenAddress := range s.config.APIListenAddresses {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%v", apiListenAddress, s.config.Port))
		if err != nil {
			log.Warnf("Failed to listen on %v:%v with error: %v", apiListenAddress, s.config.Port, err)
		} else {
			listeners = append(listeners, l)
		}
	}

	if len(listeners) == 0 {
		log.Errorf("could not listen on any endpoint")
	}

	for _, listener := range listeners {
		// server is created in a loop because each instance
		// has a handle on the underlying listener.
		server, err := s.getGRPCServer()
		if err != nil {
			return err
		}
		s.servers = append(s.servers, server)

		go func(server *grpc_go.Server, l net.Listener) {
			if err := server.Serve(l); err != nil {
				log.Fatalf("gRPC serve error: %v", err)
			}
		}(server, listener)
	}
	return nil
}

func (s *server) getGRPCServer() (*grpc_go.Server, error) {
	return grpc_go.NewServer(), nil
}

func (s *server) Close() error {
	for _, server := range s.servers {
		// This calls `Close()` on the underlying listener.
		server.GracefulStop()
	}

	return nil
}
