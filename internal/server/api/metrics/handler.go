package metrics

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NewCreateMetricHandler(service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		metricID := chi.URLParam(request, "name")
		if metricID == "" {
			http.Error(writer, "Metric ID should be filled", http.StatusNotFound)
			return
		}

		var requestDto common.MetricRequestDto
		metricType := common.MetricType(chi.URLParam(request, "type"))
		if metricType == common.Gauge {
			value, err := strconv.ParseFloat(strings.TrimSpace(chi.URLParam(request, "value")), 64)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Metric with ID = %s and type = %s has incorrect value = %f. Expected type = float64", metricID, metricType, value), http.StatusBadRequest)
				return
			}
			requestDto = common.MetricRequestDto{
				MType: metricType,
				ID:    metricID,
				Value: &value,
				Delta: nil,
			}
		}

		if metricType == common.Counter {
			value, err := strconv.ParseInt(strings.TrimSpace(chi.URLParam(request, "value")), 10, 64)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Metric with ID = %s and type = %s has incorrect value = %d. Expected type = int64", metricID, metricType, value), http.StatusBadRequest)
				return
			}
			requestDto = common.MetricRequestDto{
				MType: metricType,
				ID:    metricID,
				Value: nil,
				Delta: &value,
			}
		}

		if _, err := service.SaveMetrics(request.Context(), []common.MetricRequestDto{requestDto}); err != nil {
			var validationError *server.ValidationError
			if errors.As(err, &validationError) {
				http.Error(writer, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}

func NewCreateMetricHandlerFromJSON(service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}
		var requestDto common.MetricRequestDto

		if err := json.NewDecoder(request.Body).Decode(&requestDto); err != nil {
			logger.Error("Error decode request", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := request.Body.Close(); err != nil {
			logger.Error("Error closing response body", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		metric, err := service.SaveMetrics(request.Context(), []common.MetricRequestDto{requestDto})
		if err != nil {
			var validationError *server.ValidationError
			if errors.As(err, &validationError) {
				logger.Error("Validation was failed", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusBadRequest)
			} else {
				logger.Error("Internal server error", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(metric[0]); err != nil {
			logger.Error("Error during encode response", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func NewCreateListMetricsHandlerFromJSON(config server.Config, service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		if config.HashKey != "" {
			var requestBodyBytes bytes.Buffer
			_, err := io.Copy(&requestBodyBytes, request.Body)
			if err != nil {
				logger.Error("Error reading request body", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}

			requestBody := requestBodyBytes.Bytes()
			hash := sha256.Sum256(append(requestBody, []byte(config.HashKey)...))
			hashStr := hex.EncodeToString(hash[:])
			hashSHA256 := request.Header.Get("HashSHA256")
			if hashSHA256 == "" {
				http.Error(writer, "HashSHA256 header is missing", http.StatusBadRequest)
				return
			}

			if hashSHA256 != hashStr {
				logger.Error("Invalid HashSHA256", zap.String("header hash", hashSHA256), zap.String("calculated hash", hashStr))
				http.Error(writer, "Invalid HashSHA256", http.StatusBadRequest)
				return
			}
			request.Body = io.NopCloser(bytes.NewReader(requestBody))
		}

		var requestDtoList []common.MetricRequestDto
		if err := json.NewDecoder(request.Body).Decode(&requestDtoList); err != nil {
			logger.Error("Error decoding request", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		savedMetrics, err := service.SaveMetrics(request.Context(), requestDtoList)
		if err != nil {
			var validationError *server.ValidationError
			if errors.As(err, &validationError) {
				logger.Error("Validation was failed", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusBadRequest)
			} else {
				logger.Error("Internal server error", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		response, err := json.Marshal(savedMetrics)
		if err != nil {
			logger.Error("Error during marshal response", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		if config.HashKey != "" {
			hash := sha256.Sum256(append(response, []byte(config.HashKey)...))
			hashStr := hex.EncodeToString(hash[:])
			writer.Header().Set("HashSHA256", hashStr)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write(response); err != nil {
			logger.Error("Error during write response", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func NewFindMetricValueHandler(service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		metricName := chi.URLParam(request, "name")
		metricType := common.MetricType(chi.URLParam(request, "type"))

		if metricType != common.Gauge && metricType != common.Counter {
			http.Error(writer, fmt.Sprintf("Unsupported metric type = '%s'", metricType), http.StatusBadRequest)
			return
		}

		metric, err := service.FindOneMetric(request.Context(), metricName, metricType)
		if err != nil {
			var notFoundError *server.MetricNotFoundError
			if errors.As(err, &notFoundError) {
				http.Error(writer, err.Error(), http.StatusNotFound)
			} else {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		var value interface{}
		if metric.MType == common.Gauge {
			value = metric.Value
		} else {
			value = metric.Delta
		}

		bytes, err := json.Marshal(value)
		if err != nil {
			logger.Error("Error during marshal metric value. {}", zap.Error(err))
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err = writer.Write(bytes); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func NewFindOneMetricHandler(service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		var requestDto common.MetricRequestDto

		if err := json.NewDecoder(request.Body).Decode(&requestDto); err != nil {
			logger.Error("Error decode request", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if err := request.Body.Close(); err != nil {
			logger.Error("Error closing response body", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		if requestDto.MType != common.Gauge && requestDto.MType != common.Counter {
			http.Error(writer, fmt.Sprintf("Unsupported metric type = '%s'", requestDto.MType), http.StatusBadRequest)
			return
		}

		metric, err := service.FindOneMetric(request.Context(), requestDto.ID, requestDto.MType)
		if err != nil {
			var notFoundError *server.MetricNotFoundError
			if errors.As(err, &notFoundError) {
				http.Error(writer, err.Error(), http.StatusNotFound)
			} else {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		bytes, err := json.Marshal(metric)
		if err != nil {
			logger.Error("Error during marshal metric. {}", zap.Error(err))
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if _, err = writer.Write(bytes); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func NewFindAllMetricsHandler(service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		bytes, err := json.Marshal(service.FindAllMetrics(request.Context()))
		if err != nil {
			logger.Error("Error during marshal metric.", zap.Error(err))
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "text/html")
		if _, err = writer.Write(bytes); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func NewPingHandler(pool *pgxpool.Pool, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		if pool == nil {
			logger.Error("Connect with DB was not created")
			http.Error(writer, "Connect with DB was not created", http.StatusInternalServerError)
			return
		}

		ctx, cancelFunc := context.WithTimeout(request.Context(), 1*time.Second)
		defer cancelFunc()

		err := pool.Ping(ctx)

		if err != nil {
			logger.Error("Database is not available", zap.Error(err))
			http.Error(writer, "Database is not available", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}
