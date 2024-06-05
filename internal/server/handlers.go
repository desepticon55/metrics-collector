package server

import (
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Method '%s' is not allowed", r.Method), http.StatusBadRequest)
		return
	}
	log.Printf("URL: %s", r.URL.String())
	parts := make([]string, 3)

	copy(parts, strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/"))

	var requestDto = common.MetricRequestDto{
		Type:  common.MetricType(parts[0]),
		Name:  parts[1],
		Value: parts[2],
	}

	if requestDto.Name == "" {
		http.Error(w, "Metric name should be filled", http.StatusNotFound)
		return
	}

	validate := validator.New()
	validate.RegisterValidation("metricTypeCheck", metricTypeValidator)
	err := validate.Struct(requestDto)
	if err != nil {
		log.Printf("Request validation failed: %s", err)
		http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
	}

	switch requestDto.Type {
	case common.Gauge:
		value, err := strconv.ParseFloat(strings.TrimSpace(requestDto.Value), 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request. Gauge type has incorrect value = %s. Expected float64", requestDto.Value), http.StatusBadRequest)
		}
		handler.storage.SaveMetric(common.Metric{Name: requestDto.Name, Type: common.Gauge, Value: value})
	case common.Counter:
		value, err := strconv.ParseInt(strings.TrimSpace(requestDto.Value), 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request. Counter type has incorrect value = %s. Expected int64", requestDto.Value), http.StatusBadRequest)
		}
		metric, exists := handler.storage.GetMetric(requestDto.Name)
		if exists {
			handler.storage.SaveMetric(common.Metric{Name: requestDto.Name, Type: common.Counter, Value: metric.Value.(int64) + value})
		} else {
			handler.storage.SaveMetric(common.Metric{Name: requestDto.Name, Type: common.Counter, Value: value})
		}
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

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/find/"), "/")
	if len(parts) != 1 {
		log.Printf("Invalid URL format. URL: %s", r.URL.String())
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	metricName := parts[0]
	metric, exists := handler.storage.GetMetric(metricName)
	if !exists {
		http.Error(w, fmt.Sprintf("Metric with name '%s' was not found", metricName), http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(metric)
	if err != nil {
		log.Printf("Error during marshal mertic. %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
