package server

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewWriteMetricHandler(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name        string
		metricName  string
		metricType  string
		metricValue string
		want        want
	}{
		{
			name:        "positive test correct counter metric",
			metricName:  "some_counter_metric",
			metricType:  "counter",
			metricValue: "1",
			want: want{
				code: 200,
			},
		},
		{
			name:        "positive test with correct gauge metric",
			metricName:  "some_gauge_metric",
			metricType:  "gauge",
			metricValue: "1.83",
			want: want{
				code: 200,
			},
		},
		{
			name:        "negative test metric without metricName",
			metricType:  "counter",
			metricValue: "1",
			want: want{
				code: 404,
			},
		},
		{
			name:        "negative test metric without metricType",
			metricName:  "some_counter_metric",
			metricValue: "1",
			want: want{
				code: 400,
			},
		},
		{
			name:        "negative test metric with incorrect counter value",
			metricType:  "counter",
			metricName:  "some_counter_metric",
			metricValue: "1.34",
			want: want{
				code: 400,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/%s/%s/%s", test.metricType, test.metricName, test.metricValue), nil)
			w := httptest.NewRecorder()
			handler := WriteMetricHandler{
				storage: NewMemStorage(),
			}
			handler.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}

}

func TestNewReadMetricHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		metricName string
		want       want
	}{
		{
			name:       "test exist metric",
			metricName: "some_metric",
			want: want{
				code:        200,
				response:    `{"Name":"some_metric","Type":"counter","Value":"9"}`,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test unknown metric",
			metricName: "unknown_metric",
			want: want{
				code:        404,
				response:    `Metric with name 'unknown_metric' was not found`,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		storage := NewMemStorage()
		storage.SaveMetric(common.Metric{Name: "some_metric", Type: common.Counter, Value: "9"})
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/find/%s", test.metricName), nil)
			w := httptest.NewRecorder()
			handler := ReadMetricHandler{
				storage: storage,
			}
			handler.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			if res.StatusCode == 200 {
				assert.JSONEq(t, test.want.response, string(resBody))
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}

}
