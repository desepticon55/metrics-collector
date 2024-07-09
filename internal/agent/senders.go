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

func (s HTTPMetricsSender) SendMetrics(destination string, metrics []common.MetricRequestDto) error {
	backoff := heimdall.NewExponentialBackoff(1*time.Second, 5*time.Second, 2, 0)
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(1*time.Second),
		httpclient.WithRetrier(heimdall.NewRetrier(backoff)),
		httpclient.WithRetryCount(3),
	)

	url := fmt.Sprintf("http://%s/updates/", destination)
	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	headers.Add("Content-Encoding", "gzip")

	requestBody, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error during JSON marshaling: %v", err)
		return err
	}

	var compressedRequest bytes.Buffer
	writer := gzip.NewWriter(&compressedRequest)
	_, err = writer.Write(requestBody)
	if err != nil {
		log.Printf("Error during compressing request: %v", err)
		return err
	}
	err = writer.Close()
	if err != nil {
		log.Printf("Error closing GZIP writer: %v", err)
		return err
	}

	resp, err := client.Post(url, bytes.NewBuffer(compressedRequest.Bytes()), headers)
	if err != nil {
		log.Printf("Error during sending request: %v", err)
		return err
	}

	if err := resp.Body.Close(); err != nil {
		log.Printf("Error closing response body: %v", err)
	}

	return nil
}
