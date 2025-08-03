package rest

import (
	"net/http"

	"collector/internal/config"
	"github.com/go-chi/chi/v5"
)

type API struct {
	srv *http.Server
}

func NewAPI(router chi.Router, conf *config.ServerConfig) *API {
	return &API{
		srv: &http.Server{
			Addr:    conf.GetAddress(),
			Handler: router,
		},
	}
}
