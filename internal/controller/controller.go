package controller

import (
	"log/slog"

	"collector/internal/adapters/store"
	"collector/internal/config"
	"collector/pkg/network"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	logger *slog.Logger
	router chi.Router
	resp   *network.Response
	st     store.Store
	conf   *config.ServerConfig
}

func New(
	logger *slog.Logger,
	st store.Store,
	conf *config.ServerConfig,
	resp *network.Response,
) *Controller {
	return &Controller{logger: logger, st: st, conf: conf, router: chi.NewRouter(), resp: resp}
}
