package server

//
//import (
//	"fmt"
//	"github.com/desepticon55/metrics-collector/internal/common"
//	"github.com/go-chi/chi/v5"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"io"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//)
//
//func TestNewWriteMetricHandler(t *testing.T) {
//	type expected struct {
//		status int
//		body   string
//	}
//	tests := []struct {
//		name     string
//		path     string
//		expected expected
//	}{
//		{
//			name: "Valid counter metric",
//			path: "/update/counter/some_counter_metric/1",
//			expected: expected{
//				status: http.StatusOK,
//				body:   "",
//			},
//		},
//		{
//			name: "Valid gauge metric",
//			path: "/update/gauge/some_gauge_metric/1.33",
//			expected: expected{
//				status: http.StatusOK,
//				body:   "",
//			},
//		},
//		{
//			name: "Invalid counter metric",
//			path: "/update/counter/some_counter_metric/invalid",
//			expected: expected{
//				status: http.StatusBadRequest,
//				body:   "bad request. Counter type has incorrect value = invalid. Expected int64\n",
//			},
//		},
//		{
//			name: "Invalid gauge metric",
//			path: "/update/gauge/some_gauge_metric/invalid",
//			expected: expected{
//				status: http.StatusBadRequest,
//				body:   "bad request. Gauge type has incorrect value = invalid. Expected float64\n",
//			},
//		},
//		{
//			name: "Unsupported metric type",
//			path: "/update/unknown/some_metric/1",
//			expected: expected{
//				status: http.StatusBadRequest,
//				body:   "request validation failed: Key: 'MetricRequestDto.Type' Error:Field validation for 'Type' failed on the 'metricTypeCheck' tag\n",
//			},
//		},
//		{
//			name: "Missing value",
//			path: "/update/gauge/some_metric/",
//			expected: expected{
//				status: http.StatusNotFound,
//				body:   "404 page not found\n",
//			},
//		},
//	}
//
//	storage := NewMemStorage()
//	router := chi.NewRouter()
//	router.Method(http.MethodPost, "/update/{type}/{name}/{value}", NewWriteMetricHandler(storage))
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			req, err := http.NewRequest("POST", test.path, nil)
//			assert.NoError(t, err)
//
//			recorder := httptest.NewRecorder()
//			router.ServeHTTP(recorder, req)
//
//			assert.Equal(t, test.expected.status, recorder.Code)
//			assert.Equal(t, test.expected.body, recorder.Body.String())
//		})
//	}
//}
//
//func TestNewReadMetricHandler(t *testing.T) {
//	type expected struct {
//		code        int
//		response    string
//		contentType string
//	}
//	tests := []struct {
//		name       string
//		metricName string
//		metricType string
//		expected   expected
//	}{
//		{
//			name:       "Known metric",
//			metricName: "some_metric",
//			metricType: "counter",
//			expected: expected{
//				code:        200,
//				response:    `9`,
//				contentType: "text/plain; charset=utf-8",
//			},
//		},
//		{
//			name:       "Unknown metric",
//			metricName: "unknown_metric",
//			metricType: "counter",
//			expected: expected{
//				code:        404,
//				response:    `Metric with name 'unknown_metric' and type 'counter' was not found`,
//				contentType: "text/plain; charset=utf-8",
//			},
//		},
//	}
//	storage := NewMemStorage()
//	require.NoError(t, storage.SaveMetric(server.Metric{ID: "some_metric", Type: common.Counter, Value: 9}))
//
//	router := chi.NewRouter()
//	router.Method(http.MethodGet, "/value/{type}/{name}", NewReadMetricHandler(storage))
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			req, err := http.NewRequest("GET", fmt.Sprintf("/value/%s/%s", test.metricType, test.metricName), nil)
//			assert.NoError(t, err)
//
//			recorder := httptest.NewRecorder()
//			router.ServeHTTP(recorder, req)
//
//			res := recorder.Result()
//			assert.Equal(t, test.expected.code, recorder.Code)
//
//			defer res.Body.Close()
//			resBody, err := io.ReadAll(res.Body)
//
//			require.NoError(t, err)
//			if res.StatusCode == 200 {
//				assert.JSONEq(t, test.expected.response, string(resBody))
//			}
//			assert.Equal(t, test.expected.contentType, res.Header.Get("Content-Type"))
//		})
//	}
//
//}
//
//func TestNewReadAllMetricsHandler(t *testing.T) {
//	type expected struct {
//		code        int
//		response    string
//		contentType string
//	}
//	tests := []struct {
//		name     string
//		expected expected
//	}{
//		{
//			name: "Fetch all metrics",
//			expected: expected{
//				code:        200,
//				response:    `[{"ID":"some_counter_metric","Type":"counter","Value":9}, {"ID":"some_gauge_metric","Type":"gauge","Value":1.32}]`,
//				contentType: "text/plain; charset=utf-8",
//			},
//		},
//	}
//	storage := NewMemStorage()
//	require.NoError(t, storage.SaveMetric(server.Metric{ID: "some_counter_metric", Type: common.Counter, Value: 9}))
//	require.NoError(t, storage.SaveMetric(server.Metric{ID: "some_gauge_metric", Type: common.Gauge, Value: 1.32}))
//
//	router := chi.NewRouter()
//	router.Method(http.MethodGet, "/", NewReadAllMetricsHandler(storage))
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			req, err := http.NewRequest("GET", "/", nil)
//			assert.NoError(t, err)
//
//			recorder := httptest.NewRecorder()
//			router.ServeHTTP(recorder, req)
//
//			res := recorder.Result()
//			assert.Equal(t, test.expected.code, recorder.Code)
//
//			defer res.Body.Close()
//			resBody, err := io.ReadAll(res.Body)
//
//			require.NoError(t, err)
//			if res.StatusCode == 200 {
//				assert.JSONEq(t, test.expected.response, string(resBody))
//			}
//			assert.Equal(t, test.expected.contentType, res.Header.Get("Content-Type"))
//		})
//	}
//
//}
