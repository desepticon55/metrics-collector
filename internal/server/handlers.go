package server

import (
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type WriteMetricHandler struct {
	storage Storage
}

func NewWriteMetricHandler(s Storage) *WriteMetricHandler {
	return &WriteMetricHandler{
		storage: s,
	}
}

func (handler *WriteMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var requestDto = common.MetricRequestDto{
		Type:  common.MetricType(chi.URLParam(r, "type")),
		Name:  chi.URLParam(r, "name"),
		Value: chi.URLParam(r, "value"),
	}

	if requestDto.Name == "" {
		http.Error(w, "Metric name should be filled", http.StatusNotFound)
		return
	}

	metric, err := MapMetricRequestDtoToMetricDomainModel(requestDto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := handler.storage.SaveMetric(metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type ReadMetricHandler struct {
	storage Storage
}

func NewReadMetricHandler(s Storage) *ReadMetricHandler {
	return &ReadMetricHandler{
		storage: s,
	}
}

func (handler *ReadMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Method '%s' is not allowed", r.Method), http.StatusBadRequest)
		return
	}

	metricName := chi.URLParam(r, "name")
	metricType := common.MetricType(chi.URLParam(r, "type"))

	if metricType != common.Gauge && metricType != common.Counter {
		http.Error(w, fmt.Sprintf("Unsupported metric type = '%s'", metricType), http.StatusBadRequest)
		return
	}

	metric, exists := handler.storage.GetMetric(metricName, metricType)
	if !exists {
		http.Error(w, fmt.Sprintf("Metric with name '%s' and type '%s' was not found", metricName, metricType), http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(metric.Value)
	if err != nil {
		log.Printf("Error during marshal mertic value. %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(bytes); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

type ReadAllMetricsHandler struct {
	storage Storage
}

func NewReadAllMetricsHandler(s Storage) *ReadAllMetricsHandler {
	return &ReadAllMetricsHandler{
		storage: s,
	}
}

func (handler *ReadAllMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Method '%s' is not allowed", r.Method), http.StatusBadRequest)
		return
	}

	bytes, err := json.Marshal(handler.storage.GetAllMetrics())
	if err != nil {
		log.Printf("Error during marshal mertic. %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(bytes); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
