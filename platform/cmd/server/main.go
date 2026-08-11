package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/api/router"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/config"
	"github.com/foyez/dbaas-platform/platform/internal/infra/cnpg"
	"github.com/foyez/dbaas-platform/platform/internal/infra/k8s"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/foyez/dbaas-platform/platform/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.LogLevel, cfg.App.Env)

	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Error("failed creating kubernetes client", "error", err)
		os.Exit(1)
	}

	instanceClient := cnpg.NewClient(k8sClient)

	ctx := context.Background()

	authmw, err := auth.New(ctx, cfg.Zitadel)
	if err != nil {
		log.Error("failed to initialize auth", "error", err)
		os.Exit(1)
	}

	svc := service.NewInstanceService(instanceClient, log)
	handler := handler.NewInstanceHandler(svc, log, authmw)

	r := router.New(handler, authmw)

	setupDocs(r)

	if err := r.Run(cfg.Server.Address); err != nil {
		log.Error("failded to start API server", "error", err)
	}
}
