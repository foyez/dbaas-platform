package main

import (
	"log/slog"

	"github.com/foyez/dbaas-platform/platform/internal/api"
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
	}

	log := logger.New(cfg.App.LogLevel, cfg.App.Env)

	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Error("failed creating kubernetes client", "error", err)
		return
	}

	instanceClient := cnpg.NewClient(k8sClient)

	svc := service.NewInstanceService(instanceClient, log)
	handler := api.NewInstanceHandler(svc, log)

	r := api.NewRouter(handler)

	setupDocs(r)

	if err := r.Run(cfg.Server.Address); err != nil {
		log.Error("failded to start API server", "error", err)
	}
}
