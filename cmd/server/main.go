package main

import (
	"context"
	"encoding/json"
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
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	logger, err := common.NewLogger()
	if err != nil {
		log.Fatal("Error during initialise logger", err)
	}
	defer logger.Sync()

	config := extractConfig(logger)
	flag.Parse()
	v := initValidator()
	mapper := initMapper(v)

	var metricsService metricsServices.Service
	pool, err := createConnectionPool(context.Background(), config.DatabaseConnString)
	if err != nil {
		logger.Debug("Run with memory/file storage")
		storage := memory.New(config.FileStoragePath, config.Restore, time.Duration(config.StoreInterval)*time.Second)
		metricsService = metricsServices.New(storage, mapper, server.NewRetrier(3, 1*time.Second, 5*time.Second))
	} else {
		logger.Debug("Run with Postgres storage")
		runMigrations(config.DatabaseConnString, logger)
		storage := postgres.New(pool, logger)
		metricsService = metricsServices.New(storage, mapper, server.NewRetrier(3, 1*time.Second, 5*time.Second))
	}

	logger.Info("Current config:", zap.String("config", config.String()))

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(customMiddleware.LoggingMiddleware(logger))
	router.Use(customMiddleware.CompressingMiddleware())
	router.Use(customMiddleware.DecompressingMiddleware())
	router.Use(customMiddleware.TrustedSubnetMiddleware(config.TrustedSubnet))

	router.Method(http.MethodGet, "/", metricsApi.NewFindAllMetricsHandler(metricsService, logger))
	router.Method(http.MethodGet, "/ping", metricsApi.NewPingHandler(pool, logger))
	router.Method(http.MethodGet, "/value/{type}/{name}", metricsApi.NewFindMetricValueHandler(metricsService, logger))
	router.Method(http.MethodPost, "/value/", metricsApi.NewFindOneMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/{type}/{name}/{value}", metricsApi.NewCreateMetricHandler(metricsService, logger))
	router.Method(http.MethodPost, "/update/", metricsApi.NewCreateMetricHandlerFromJSON(metricsService, logger))
	router.Method(http.MethodPost, "/updates/", metricsApi.NewCreateListMetricsHandlerFromJSON(config, metricsService, logger))

	if config.EnabledHTTPS {
		e := http.ListenAndServeTLS(config.ServerAddress, "./cmd/cert/server.crt", config.CryptoKey, router)
		if e != nil {
			logger.Error("Error during start server", zap.Error(e))
		}
	} else {
		http.ListenAndServe(config.ServerAddress, router)
	}
}

func extractConfig(logger *zap.Logger) server.Config {
	return server.CreateConfig(logger, func(filePath string) (server.Config, error) {
		var config server.Config
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return config, fmt.Errorf("could not read config file: %w", err)
		}

		err = json.Unmarshal(fileContent, &config)
		if err != nil {
			return config, fmt.Errorf("could not unmarshal config JSON: %w", err)
		}

		return config, nil
	})
}

func initValidator() *validator.Validate {
	v := validator.New()
	v.RegisterStructValidation(server.MetricValidator, common.MetricRequestDto{})
	return v
}

func initMapper(v *validator.Validate) metricsMappers.Mapper {
	return metricsMappers.NewMapper(v)
}

func createConnectionPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing database config: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return pool, nil
}

func runMigrations(connectionString string, logger *zap.Logger) {
	databaseConfig, err := pgx.ParseConfig(connectionString)
	if err != nil {
		logger.Error("Error during parse database URL", zap.Error(err))
		return
	}
	db := stdlib.OpenDB(*databaseConfig)
	defer db.Close()

	goose.SetDialect("postgres")
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Error("Error during run database migrations", zap.Error(err))
	}
}

func loadConfigFromFile(path string) (server.Config, error) {
	var config server.Config
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("could not read config file: %w", err)
	}

	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return config, fmt.Errorf("could not unmarshal config JSON: %w", err)
	}

	return config, nil
}
