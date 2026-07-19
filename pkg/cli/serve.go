package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kommodity-io/kommodity/pkg/libkapi"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/nicklasfrahm/kontinuum/pkg/config"
	"github.com/nicklasfrahm/kontinuum/pkg/logging"
	"github.com/nicklasfrahm/kontinuum/pkg/ui"
)

const shutdownTimeout = 10 * time.Second

// NewServeCmd builds the serve command, which starts the Kubernetes-style
// API server.
func NewServeCmd() *cobra.Command {
	defaults := config.Defaults()

	var addr = defaults.Server.Addr

	var storage = defaults.Server.Storage

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Kubernetes-style API server",
		// Runtime errors (listener failures, storage errors) shouldn't print
		// the command usage alongside the error.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd, addr, storage)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", defaults.Server.Addr,
		"Listener address (e.g. \":8080\")")
	cmd.Flags().StringVar(&storage, "storage", defaults.Server.Storage,
		"Storage connection string (e.g. sqlite://kontinuum.db, postgres://...)")

	return cmd
}

// runServe loads config, builds the libkapi server, and runs it until a
// signal is received or an unrecoverable error occurs.
func runServe(cmd *cobra.Command, addr string, storage string) error {
	cfg, logger, err := loadServeConfig(cmd, addr, storage)
	if err != nil {
		return err
	}

	// sigChan catches SIGINT and SIGTERM so we can log which signal was
	// received before initiating shutdown.
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := buildServer(cfg, logger)
	if err != nil {
		return err
	}

	logger.Info("Kontinuum starting", "addr", cfg.Server.Addr, "storage", cfg.Server.Storage)

	// Run the server in a goroutine so we can watch for signals on the
	// main goroutine and log which signal was received.
	serveErr := make(chan error, 1)

	go func() {
		serveErr <- server.ListenAndServe(ctx)
	}()

	sig := <-sigChan

	logger.Info("Received signal, shutting down", "signal", sig.String())

	cancel()

	err = shutdownServer(server, logger)
	if err != nil {
		<-serveErr

		return err
	}

	err = <-serveErr
	if err != nil {
		return fmt.Errorf("server exited with error: %w", err)
	}

	return nil
}

// loadServeConfig loads config from environment variables, applies flag
// overrides, and creates the logger.
func loadServeConfig(cmd *cobra.Command, addr string, storage string) (*config.Config, *slog.Logger, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Flags override config (env vars) when explicitly set.
	if cmd.Flags().Changed("addr") {
		cfg.Server.Addr = addr
	}

	if cmd.Flags().Changed("storage") {
		cfg.Server.Storage = storage
	}

	level, err := logging.ParseLevel(cfg.Log.Level)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	format, err := logging.ParseFormat(cfg.Log.Format)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse log format: %w", err)
	}

	logger := logging.New(level, format, os.Stdout)

	return cfg, logger, nil
}

// buildServer creates the libkapi server with custom handlers.
func buildServer(cfg *config.Config, logger *slog.Logger) (*libkapi.Server, error) {
	clientset, err := kubernetes.NewForConfig(&rest.Config{Host: localBaseURL(cfg.Server.Addr)})
	if err != nil {
		return nil, fmt.Errorf("failed to build in-process Kubernetes client: %w", err)
	}

	uiRouter := ui.NewRouter(clientset.CoreV1().Namespaces(), version, config.Redact(*cfg))

	kapiCfg := libkapi.Config{
		Addr:    cfg.Server.Addr,
		Storage: cfg.Server.Storage,
		Logger:  logger,
		Handlers: []libkapi.HTTPHandlerFactory{
			customHandlers(uiRouter),
		},
	}

	// Storage is resolved against a background context so the backend
	// is only torn down by Server.Shutdown, not by the signal context
	// that drives ListenAndServe.
	server, err := libkapi.New(context.Background(), kapiCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build server: %w", err)
	}

	return server, nil
}

// shutdownServer gracefully stops the HTTP listener, the apiserver's
// background run loop, and the storage backend.
func shutdownServer(server *libkapi.Server, logger *slog.Logger) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil && !errors.Is(err, libkapi.ErrServerNotStarted) {
		logger.Error("Graceful shutdown failed", "error", err)

		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

// customHandlers mounts example routes and the /app UI alongside the built
// API server. Any request that does not match a registered route falls
// through to the Kubernetes API server's own handler.
func customHandlers(uiRouter *ui.Router) libkapi.HTTPHandlerFactory {
	return func(mux *http.ServeMux) error {
		mux.HandleFunc("GET /kontinuum/info", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"kontinuum","kind":"kubernetes-style-api"}`))
		})

		uiRouter.RegisterRoutes(mux)

		return nil
	}
}

// localBaseURL derives the loopback URL the in-process Kubernetes client
// uses to reach the server the UI is mounted on, e.g. ":8080" ->
// "http://127.0.0.1:8080". A missing or wildcard host (":8080",
// "0.0.0.0:8080") is rewritten to the loopback address since the listener
// isn't guaranteed to be reachable there.
func localBaseURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}

	return "http://" + net.JoinHostPort(host, port)
}
