package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	metrics2 "github.com/desepticon55/metrics-collector/proto/metrics"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type MetricsSender interface {
	SendMetrics(destination string, metrics []common.MetricRequestDto) error
}

type HTTPMetricsSender struct {
	config Config
}

func NewHTTPSender(config Config) MetricsSender {
	return HTTPMetricsSender{config: config}
}

func (s HTTPMetricsSender) SendMetrics(url string, metrics []common.MetricRequestDto) error {
	backoff := heimdall.NewExponentialBackoff(1*time.Second, 5*time.Second, 2, 0)

	var transport *http.Transport
	if s.config.EnabledHTTPS {
		certPool := x509.NewCertPool()

		serverCert, err := os.ReadFile(s.config.CryptoKey)
		if err != nil {
			log.Printf("Failed to load server certificate: %v", err)
			return err
		}

		if ok := certPool.AppendCertsFromPEM(serverCert); !ok {
			log.Printf("Failed to append server certificate to trust pool")
			return fmt.Errorf("could not append server certificate")
		}

		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: true,
			},
		}
	} else {
		transport = &http.Transport{}
	}

	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(1*time.Second),
		httpclient.WithRetrier(heimdall.NewRetrier(backoff)),
		httpclient.WithRetryCount(3),
		httpclient.WithHTTPClient(&http.Client{
			Transport: transport,
			Timeout:   1 * time.Second,
		}),
	)

	hostIP, err := getCurrentIP()
	if err != nil {
		return err
	}

	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	headers.Add("Content-Encoding", "gzip")
	headers.Add("X-Real-IP", hostIP)

	requestBody, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error during JSON marshaling: %v", err)
		return err
	}

	if s.config.HashKey != "" {
		hash := sha256.Sum256(append(requestBody, []byte(s.config.HashKey)...))
		hashStr := hex.EncodeToString(hash[:])
		headers.Add("HashSHA256", hashStr)
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

type GRPCMetricsSender struct {
	config Config
}

func NewGRPCSender(config Config) MetricsSender {
	return GRPCMetricsSender{config: config}
}

func (s GRPCMetricsSender) SendMetrics(url string, metrics []common.MetricRequestDto) error {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
		return err
	}
	defer conn.Close()

	client := metrics2.NewMetricsServiceClient(conn)
	hostIP, err := getCurrentIP()
	if err != nil {
		return err
	}

	var protoMetrics []*metrics2.Metric
	for _, m := range metrics {
		protoMetric := &metrics2.Metric{
			Id:   m.ID,
			Type: string(m.MType),
		}

		if m.Delta != nil {
			protoMetric.Delta = *m.Delta
		}

		if m.Value != nil {
			protoMetric.Value = *m.Value
		}

		protoMetrics = append(protoMetrics, protoMetric)
	}

	hash := ""
	if s.config.HashKey != "" {
		requestBody, _ := json.Marshal(protoMetrics)
		hashSum := sha256.Sum256(append(requestBody, []byte(s.config.HashKey)...))
		hash = hex.EncodeToString(hashSum[:])
	}

	_, err = client.SendMetrics(context.Background(), &metrics2.MetricsRequest{
		Metrics: protoMetrics,
		Ip:      hostIP,
		Hash:    hash,
	})

	return err
}

func getCurrentIP() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addresses {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid IP address found")
}
