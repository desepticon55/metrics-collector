package main

import (
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func main() {
	storage := server.NewMemStorage()

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Method(http.MethodGet, "/", server.NewReadAllMetricsHandler(storage))
	router.Method(http.MethodGet, "/value/{type}/{name}", server.NewReadMetricHandler(storage))
	router.Method(http.MethodPost, "/update/{type}/{name}/{value}", server.NewWriteMetricHandler(storage))

	http.ListenAndServe(":8080", router)
}
