package main

import (
	"github.com/desepticon55/metrics-collector/internal/agent"
	"github.com/desepticon55/metrics-collector/internal/common"
	"log"
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	var metrics []common.Metric

	provider := &agent.RuntimeMetricProvider{}
	go func() {
		for range time.Tick(2 * time.Second) {
			m := provider.GetMetrics()
			mu.Lock()
			metrics = append(metrics, m...)
			mu.Unlock()
		}
	}()

	sender := &agent.HTTPMetricsSender{}
	for range time.Tick(10 * time.Second) {
		mu.Lock()
		err := sender.SendMetrics(metrics)
		if err != nil {
			log.Printf("Error during  metrics: %s", err)
		} else {
			log.Println("Metrics sent successfully")
		}
		metrics = nil
		mu.Unlock()
	}
}
