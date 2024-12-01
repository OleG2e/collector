package main

import (
	"github.com/OleG2e/collector/internal/controller"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`POST /update/counter/{metric}/{value}`, controller.UpdateCounter())
	mux.HandleFunc(`POST /update/gauge/{metric}/{value}`, controller.UpdateGauge())
	mux.Handle(`POST /update/counter/`, http.NotFoundHandler())
	mux.Handle(`POST /update/gauge/`, http.NotFoundHandler())
	mux.Handle("/", controller.BadRequestHandler())

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
