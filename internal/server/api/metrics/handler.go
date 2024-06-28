package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
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

		if _, err := service.SaveMetric(requestDto); err != nil {
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

		metric, err := service.SaveMetric(requestDto)
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
		if err := json.NewEncoder(writer).Encode(metric); err != nil {
			logger.Error("Error during encode response", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		writer.WriteHeader(http.StatusOK)
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

		metric, err := service.FindOneMetric(metricName, metricType)
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

		metric, err := service.FindOneMetric(requestDto.ID, requestDto.MType)
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

func NewFinAllMetricsHandler(service metricsService, logger *zap.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		bytes, err := json.Marshal(service.FindAllMetrics())
		if err != nil {
			logger.Error("Error during marshal metric.", zap.Error(err))
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		if _, err = writer.Write(bytes); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
	}
}
