package manager

import (
	"context"
	"github.com/woojoong88/namespace-watchers/modules/informer/internal/kubernetes"
	"github.com/woojoong88/namespace-watchers/modules/informer/internal/probe"
	"time"
)

const (
	// EnvKeyExcludedNamespaces is the environment variable key for excluded namespaces
	EnvKeyExcludedNamespaces = "EXCLUDED_NAMESPACES"
)

// Config is the configuration for the Manager
type Config struct {
	// ProbePort is the port for readiness and liveness probe
	ProbePort int
	// ExcludedNamespaces is the list of namespaces to be excluded from watching
	ExcludedNamespaces []string
}

// Manager is the interface for the manager struct
type Manager interface {
	// Run starts the manager to watch namespaces
	Run(ctx context.Context) error
}

// manager is the struct that implements the Manager interface, which includes configs and health server for readiness and liveness probe
type manager struct {
	// config is the configuration for the Manager
	config Config
	// healthServer is the health server for readiness and liveness probe
	healthServer probe.Server
}

// NewManager creates a new Manager with the given configuration
func NewManager(config Config) Manager {
	return &manager{
		config:       config,
		healthServer: probe.NewServer(config.ProbePort),
	}
}

// Run starts the manager to watch namespaces
func (m manager) Run(ctx context.Context) error {
	// start health server for readiness and liveness probe
	go m.healthServer.StartHealthServers(ctx)

	// set ready to true for readiness probe
	m.healthServer.SetReady(true)
	defer m.healthServer.SetReady(false)

	// start watching namespaces
	// it is used for checking if the namespace is new or pre-existing namespace
	startTime := time.Now()

	// watch namespaces
	return kubernetes.WatchNamespace(ctx, m.config.ExcludedNamespaces, startTime)
}
