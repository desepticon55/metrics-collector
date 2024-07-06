package main

import (
	"context"
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
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
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

	connect, err := pgx.Connect(context.Background(), config.DatabaseConnString)
	if err != nil {
		logger.Error("Error during connect to database", zap.Error(err))
	}
	defer connect.Close(context.Background())

	storage := memory.New(config.FileStoragePath, config.Restore, time.Duration(config.StoreInterval)*time.Second)
	mapper := metricsMappers.NewMapper()
	metricsService := metricsServices.New(storage, mapper)

	fmt.Println("Address:", config.ServerAddress)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(customMiddleware.LoggingMiddleware(logger))
	router.Use(customMiddleware.CompressingMiddleware())
	router.Use(customMiddleware.DecompressingMiddleware())

	router.Method(http.MethodGet, "/", metricsApi.NewFindAllMetricsHandler(metricsService, logger))
	router.Method(http.MethodGet, "/ping", metricsApi.NewPingHandler(connect, logger))
	router.Method(http.MethodGet, "/value/{type}/{name}", metricsApi.NewFindMetricValueHandler(metricsService, logger))
	router.Method(http.MethodPost, "/value/", metricsApi.NewFindOneMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/{type}/{name}/{value}", metricsApi.NewCreateMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/", metricsApi.NewCreateMetricHandlerFromJSON(metricsService, logger))

	http.ListenAndServe(config.ServerAddress, router)
}
