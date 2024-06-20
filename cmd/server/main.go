package main

import (
	"flag"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	metricsApi "github.com/desepticon55/metrics-collector/internal/server/api/metrics"
	customMiddleware "github.com/desepticon55/metrics-collector/internal/server/api/middleware"
	metricsMappers "github.com/desepticon55/metrics-collector/internal/server/mapper/metrics"
	metricsServices "github.com/desepticon55/metrics-collector/internal/server/service/metrics"
	"github.com/desepticon55/metrics-collector/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func main() {
	logger, err := common.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	config := server.ParseConfig()
	flag.Parse()

	storage := memory.New()
	mapper := metricsMappers.NewMapper()
	metricsService := metricsServices.New(storage, mapper)

	fmt.Println("Address:", config.ServerAddress)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(customMiddleware.LoggingMiddleware(logger))

	router.Method(http.MethodGet, "/", metricsApi.NewFinAllMetricsHandler(metricsService, logger))
	router.Method(http.MethodGet, "/value/{type}/{name}", metricsApi.NewFinOneMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/{type}/{name}/{value}", metricsApi.NewCreateMetricHandler(metricsService, logger))

	http.ListenAndServe(config.ServerAddress, router)
}
