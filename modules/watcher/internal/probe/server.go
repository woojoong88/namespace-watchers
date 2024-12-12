package probe

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

const (
	// livenessPath is the path for liveness probe
	livenessPath = "/healthz"
	// readinessPath is the path for readiness probe
	readinessPath = "/readyz"
)

// Server is the interface for the health server
type Server interface {
	// StartHealthServers starts the health server for readiness and liveness probe
	StartHealthServers(ctx context.Context)
	// SetReady sets the readiness of the server
	SetReady(isReady bool)
}

// server is the struct that implements the Server interface
type server struct {
	// mu is the mutex for the readiness
	mu sync.RWMutex
	// port is the port for the health server
	port int
	// isReady is the readiness of the server
	isReady bool
}

// NewServer creates a new Server with the given port
func NewServer(port int) Server {
	return &server{
		port:    port,
		isReady: false,
	}
}

// SetReady sets the readiness of the server
func (s *server) SetReady(isReady bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isReady = isReady
}

// StartHealthServers starts the health server for readiness and liveness probe
func (s *server) StartHealthServers(_ context.Context) {
	// set liveness probe handler
	http.HandleFunc(livenessPath, s.handleLiveness)
	// set readiness probe handler
	http.HandleFunc(readinessPath, s.handleReadiness)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.port),
	}

	// open health server
	if err := server.ListenAndServe(); err != nil {
		logrus.Error(err, "error starting health server")
	}

	// if health server stopped unexpectedly, panic to stop the service and then restart it by Kubernetes
	panic("health server stopped unexpectedly")
}

// handleLiveness is the handler for liveness probe
func (s *server) handleLiveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadiness is the handler for readiness probe
func (s *server) handleReadiness(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Unavailable"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
