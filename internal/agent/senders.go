package agent

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
	"log"
	"net/http"
	"time"
)

type MetricsSender interface {
	SendMetrics(metrics []common.Metric) error
}

type HTTPMetricsSender struct {
}

func (s *HTTPMetricsSender) SendMetrics(address string, metrics []common.Metric) error {
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(1*time.Second),
		httpclient.WithRetrier(heimdall.NewRetrier(heimdall.NewConstantBackoff(2*time.Second, 5*time.Second))),
		httpclient.WithRetryCount(3),
	)

	for _, metric := range metrics {
		url := fmt.Sprintf("http://%s/update/%s/%s/%s", address, metric.Type, metric.Name, metric.Value)
		headers := make(http.Header)
		headers.Add("Content-Type", "text/plain")

		resp, err := client.Post(url, nil, headers)
		if err != nil {
			return err
		}

		if resp != nil {
			func() {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						log.Printf("Error closing response body: %v", err)
					}
				}()
			}()
		}
	}

	return nil
}
