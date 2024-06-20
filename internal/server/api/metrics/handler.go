package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func NewCreateMetricHandler(service metricsService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var requestDto = common.MetricRequestDto{
			Type:  common.MetricType(chi.URLParam(request, "type")),
			Name:  chi.URLParam(request, "name"),
			Value: chi.URLParam(request, "value"),
		}

		if requestDto.Name == "" {
			http.Error(writer, "Metric name should be filled", http.StatusNotFound)
			return
		}

		if err := service.SaveMetric(requestDto); err != nil {
			var validationError *server.ValidationError
			var incorrectValueError *server.IncorrectMetricValueError
			if errors.As(err, &validationError) || errors.As(err, &incorrectValueError) {
				http.Error(writer, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		writer.WriteHeader(http.StatusOK)
	}
}

func NewFinOneMetricHandler(service metricsService) http.HandlerFunc {
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

		bytes, err := json.Marshal(metric.Value)
		if err != nil {
			log.Printf("Error during marshal mertic value. %s", err)
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		if _, err = writer.Write(bytes); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func NewFinAllMetricsHandler(service metricsService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, fmt.Sprintf("Method '%s' is not allowed", request.Method), http.StatusBadRequest)
			return
		}

		bytes, err := json.Marshal(service.FindAllMetrics())
		if err != nil {
			log.Printf("Error during marshal mertic. %s", err)
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}

		if _, err = writer.Write(bytes); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
