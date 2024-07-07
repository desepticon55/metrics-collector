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
	"github.com/desepticon55/metrics-collector/internal/server/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
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

	databaseConfig, err := pgx.ParseConfig(config.DatabaseConnString)
	if err != nil {
		logger.Error("Error during parse database URL", zap.Error(err))
	}
	runMigrations(databaseConfig, logger)
	connection := createConnection(databaseConfig, logger)
	defer connection.Close(context.Background())

	mapper := metricsMappers.NewMapper()
	var metricsService metricsServices.Service
	if connection != nil {
		storage := postgres.New(connection, logger)
		metricsService = metricsServices.New(storage, mapper)
	} else {
		storage := memory.New(config.FileStoragePath, config.Restore, time.Duration(config.StoreInterval)*time.Second)
		metricsService = metricsServices.New(storage, mapper)
	}

	fmt.Println("Config:", config)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(customMiddleware.LoggingMiddleware(logger))
	router.Use(customMiddleware.CompressingMiddleware())
	router.Use(customMiddleware.DecompressingMiddleware())

	router.Method(http.MethodGet, "/", metricsApi.NewFindAllMetricsHandler(metricsService, logger))
	router.Method(http.MethodGet, "/ping", metricsApi.NewPingHandler(connection, logger))
	router.Method(http.MethodGet, "/value/{type}/{name}", metricsApi.NewFindMetricValueHandler(metricsService, logger))
	router.Method(http.MethodPost, "/value/", metricsApi.NewFindOneMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/{type}/{name}/{value}", metricsApi.NewCreateMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/", metricsApi.NewCreateMetricHandlerFromJSON(metricsService, logger))
	router.Method(http.MethodPost, "/updates/", metricsApi.NewCreateListMetricsHandlerFromJSON(metricsService, logger))

	http.ListenAndServe(config.ServerAddress, router)
}

func createConnection(databaseConfig *pgx.ConnConfig, logger *zap.Logger) *pgx.Conn {
	connect, err := pgx.ConnectConfig(context.Background(), databaseConfig)
	if err != nil {
		logger.Error("Error during connect to database", zap.Error(err))
		return nil
	}
	return connect
}

func runMigrations(databaseConfig *pgx.ConnConfig, logger *zap.Logger) {
	db := stdlib.OpenDB(*databaseConfig)
	defer db.Close()

	goose.SetDialect("postgres")
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Error("Error during run database migrations", zap.Error(err))
	}
}
