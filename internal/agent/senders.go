package agent

import (
	"bytes"
	"compress/gzip"
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
		headers.Add("Content-Encoding", "gzip")

		request, err := json.Marshal(metric)
		if err != nil {
			return err
		}

		var compressedRequest bytes.Buffer
		writer := gzip.NewWriter(&compressedRequest)
		_, err = writer.Write(request)
		if err != nil {
			log.Printf("Error during compressed request: %v", err)
			return err
		}
		err = writer.Close()
		if err != nil {
			log.Printf("Error closing GZIP writer: %v", err)
			return err
		}

		resp, err := client.Post(url, bytes.NewBuffer(compressedRequest.Bytes()), headers)
		if err != nil {
			log.Printf("Error during send request: %v", err)
			return err
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}

	return nil
}
