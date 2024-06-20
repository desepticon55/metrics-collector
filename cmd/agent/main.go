package main

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/agent"
	"github.com/desepticon55/metrics-collector/internal/common"
	"log"
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	var metrics []common.Metric
	config := agent.GetConfig()

	fmt.Println("Address:", config.ServerAddress)
	fmt.Println("Report interval:", config.ReportInterval)
	fmt.Println("Poll interval:", config.PollInterval)

	provider := &agent.RuntimeMetricProvider{}
	go func() {
		for range time.Tick(time.Duration(config.PollInterval) * time.Second) {
			m := provider.GetMetrics()
			mu.Lock()
			metrics = append(metrics, m...)
			mu.Unlock()
		}
	}()

	sender := agent.HTTPMetricsSender{}
	for range time.Tick(time.Duration(config.ReportInterval) * time.Second) {
		mu.Lock()
		err := sender.SendMetrics(config.ServerAddress, metrics)
		if err != nil {
			log.Printf("Error during send metrics: %s", err)
		} else {
			log.Println("Metrics sent successfully")
		}
		metrics = nil
		mu.Unlock()
	}
}
