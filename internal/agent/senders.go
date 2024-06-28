package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
	"log"
	"net/http"
	"time"
)

type MetricsSender interface {
	SendMetrics(destination string, metrics []common.MetricRequestDto) error
}

type HTTPMetricsSender struct {
}

func (HTTPMetricsSender) SendMetrics(destination string, metrics []common.MetricRequestDto) error {
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(1*time.Second),
		httpclient.WithRetrier(heimdall.NewRetrier(heimdall.NewConstantBackoff(2*time.Second, 5*time.Second))),
		httpclient.WithRetryCount(3),
	)

	for _, metric := range metrics {
		url := fmt.Sprintf("http://%s/update/", destination)
		headers := make(http.Header)
		headers.Add("Content-Type", "application/json")

		request, err := json.Marshal(metric)
		if err != nil {
			return err
		}

		resp, err := client.Post(url, bytes.NewBuffer(request), headers)
		if err != nil {
			return err
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}

	return nil
}
