package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/api/router"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/config"
	"github.com/foyez/dbaas-platform/platform/internal/infra/cnpg"
	"github.com/foyez/dbaas-platform/platform/internal/infra/k8s"
	"github.com/foyez/dbaas-platform/platform/internal/infra/loki"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/foyez/dbaas-platform/platform/internal/observability"
	"github.com/foyez/dbaas-platform/platform/internal/observability/audit"
	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/prometheus/client_golang/prometheus"
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

	log := logger.New(cfg.App.LogLevel, cfg.App.Env)

	metrics := observability.NewMetrics()
	registry := prometheus.NewRegistry()
	if err := metrics.Register(registry); err != nil {
		return fmt.Errorf("register metrics: %w", err)
	}

	k8sCfg, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("load kubernetes config: %w", err)
	}

	k8sClient, err := k8s.NewClient(k8sCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	instanceClient := cnpg.NewClient(k8sClient)
	lokiClient := loki.NewClient(cfg.Server.LokiURL)

	ctx := context.Background()

	authmw, err := auth.New(ctx, cfg.Zitadel)
	if err != nil {
		return fmt.Errorf("initialize auth: %w", err)
	}

	svc := service.NewInstanceService(instanceClient, lokiClient, log)
	handler := handler.NewInstanceHandler(svc, log, authmw)
	auditLogger := audit.NewLogger(log)

	r := router.New(
		handler,
		cfg.Server,
		authmw,
		metrics,
		registry,
		auditLogger,
	)

	setupDocs(r)

	if err := r.Run(cfg.Server.Address); err != nil {
		return fmt.Errorf("start API server: %w", err)
	}

	return nil
}
