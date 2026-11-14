package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/api/router"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/config"
	"github.com/foyez/dbaas-platform/platform/internal/infra/cnpg"
	"github.com/foyez/dbaas-platform/platform/internal/infra/k8s"
	"github.com/foyez/dbaas-platform/platform/internal/infra/loki"
	"github.com/foyez/dbaas-platform/platform/internal/infra/pgquery"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/foyez/dbaas-platform/platform/internal/observability"
	"github.com/foyez/dbaas-platform/platform/internal/observability/audit"
	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	log := logger.New(cfg.App.LogLevel, cfg.App.LogFormat, cfg.App.Env)
	// Libraries that log through slog's default (e.g. zitadel-go) get the
	// same structured output.
	slog.SetDefault(log)

	// Stop on SIGTERM (Kubernetes) or Ctrl+C, so in-flight requests finish
	// and buffered spans are flushed.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := observability.SetupTracing(ctx)
	if err != nil {
		return fmt.Errorf("set up tracing: %w", err)
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(flushCtx); err != nil {
			log.Error("flush traces", "error", err)
		}
	}()

	metrics := observability.NewMetrics()
	registry := prometheus.NewRegistry()
	if err := metrics.Register(registry); err != nil {
		return fmt.Errorf("register metrics: %w", err)
	}

	k8sCfg, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("load kubernetes config: %w", err)
	}
	// A client span for every Kubernetes API call.
	k8sCfg.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return otelhttp.NewTransport(rt)
	})

	k8sClient, err := k8s.NewClient(k8sCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	instanceClient := cnpg.NewClient(k8sClient, cfg.Server.PublicDBHost)
	lokiClient := loki.NewClient(cfg.Server.LokiURL)
	querier := pgquery.NewPool()
	defer querier.Close()

	authmw, err := auth.New(ctx, cfg.Zitadel)
	if err != nil {
		return fmt.Errorf("initialize auth: %w", err)
	}

	svc := service.NewInstanceService(instanceClient, querier, lokiClient, log)
	handler := handler.NewInstanceHandler(svc, log, authmw, metrics)
	auditLogger := audit.NewLogger(log)

	r := router.New(
		handler,
		log,
		cfg.Server,
		authmw,
		metrics,
		registry,
		auditLogger,
	)

	srv := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("API listening", "address", cfg.Server.Address)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("start API server: %w", err)
	case <-ctx.Done():
	}

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down API server: %w", err)
	}

	return nil
}
