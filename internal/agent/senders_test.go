package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMetricsSender_SendMetrics(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		if r.Header.Get("HashSHA256") != "" {
			body, _ := io.ReadAll(r.Body)
			defer r.Body.Close()

			gz, err := gzip.NewReader(bytes.NewBuffer(body))
			assert.NoError(t, err)
			defer gz.Close()

			decompressedBody, _ := io.ReadAll(gz)
			expectedMetrics, _ := json.Marshal(getSampleMetrics())

			assert.Equal(t, expectedMetrics, decompressedBody)

			hash := sha256.Sum256(append(decompressedBody, []byte("test_key")...))
			expectedHash := hex.EncodeToString(hash[:])
			assert.Equal(t, expectedHash, r.Header.Get("HashSHA256"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() {
		testServer.Close()
	})

	config := Config{
		HashKey: "test_key",
	}

	sender := HTTPMetricsSender{config: config}

	metrics := getSampleMetrics()

	err := sender.SendMetrics(testServer.URL, metrics)
	assert.NoError(t, err)
}

func getSampleMetrics() []common.MetricRequestDto {
	value := float64(123.45)
	return []common.MetricRequestDto{
		{
			ID:    "TestGauge",
			MType: common.Gauge,
			Value: &value,
		},
		{
			ID:    "TestCounter",
			MType: common.Counter,
			Delta: new(int64),
		},
	}
}
