package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
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

func New(config Config) MetricsSender {
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
