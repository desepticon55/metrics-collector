package agent

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"net/http"
)

type MetricsSender interface {
	SendMetrics(metrics []common.Metric) error
}

type HTTPMetricsSender struct {
}

func (s *HTTPMetricsSender) SendMetrics(metrics []common.Metric) error {
	for _, metric := range metrics {
		url := fmt.Sprintf("%s/update/%s/%s/%s", "http://localhost:8080", metric.Type, metric.Name, metric.Value)
		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error during send metric: %s", resp.Status)
		}
	}
	return nil
}
