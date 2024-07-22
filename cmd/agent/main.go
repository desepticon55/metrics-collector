package main

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/agent"
	"github.com/desepticon55/metrics-collector/internal/common"
	"go.uber.org/zap"
	"sync"
	"time"
)

func main() {
	logger, err := common.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	var mu sync.Mutex
	var metrics []common.MetricRequestDto
	config := agent.GetConfig()

	fmt.Println("Config:", config)

	provider := &agent.RuntimeMetricProvider{}
	go func() {
		for range time.Tick(time.Duration(config.PollInterval) * time.Second) {
			m := provider.GetMetrics()
			mu.Lock()
			metrics = append(metrics, m...)
			mu.Unlock()
		}
	}()

	sender := agent.New(config)

	for range time.Tick(time.Duration(config.ReportInterval) * time.Second) {
		mu.Lock()
		err := sender.SendMetrics(config.ServerAddress, metrics)
		if err != nil {
			logger.Error("Error during send metrics", zap.Error(err))
		} else {
			logger.Info("Metrics sent successfully")
		}
		metrics = nil
		mu.Unlock()
	}
}
