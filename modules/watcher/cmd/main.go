package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/woojoong88/namespace-watchers/modules/watcher/internal/manager"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	// get probe port from flag - for readiness and liveness probe
	probePort := flag.Int("health-probe-bind-address", 8081, "The port the probe endpoint binds to.")

	// get excluded namespaces from environment variable
	var excludedNamespaces []string
	ns := os.Getenv(manager.EnvKeyExcludedNamespaces)
	if ns != "" {
		excludedNamespaces = strings.Split(ns, ",")
	}

	ctx := context.Background()
	logrus.Info("Starting watcher service")

	// create manager config to pass configs to Manager
	cfg := manager.Config{
		ProbePort:          *probePort,
		ExcludedNamespaces: excludedNamespaces,
	}

	// to handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// run Manager in a goroutine - keep retrying even if error happens
	go func() {
		// todo: apply back-off strategy
		for {
			mgr := manager.NewManager(cfg)
			err := mgr.Run(ctx)
			if err != nil {
				logrus.Error(err, " error running manager")
			}
		}
	}()

	// wait for signal to shutdown
	<-signalChan
	logrus.Info("Shutting down watcher service")
}
