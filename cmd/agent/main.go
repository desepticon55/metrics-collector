package main

import (
	"flag"
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
	pollInterval := flag.Int("p", 2, "Server address")
	reportInterval := flag.Int("r", 10, "Server address")
	address := flag.String("a", "localhost:8080", "Server address")
	flag.Parse()

	fmt.Println("Address:", *address)
	fmt.Println("Report interval:", *reportInterval)
	fmt.Println("Poll interval:", *pollInterval)

	provider := &agent.RuntimeMetricProvider{}
	go func() {
		for range time.Tick(time.Duration(*pollInterval) * time.Second) {
			m := provider.GetMetrics()
			mu.Lock()
			metrics = append(metrics, m...)
			mu.Unlock()
		}
	}()

	sender := &agent.HTTPMetricsSender{}
	for range time.Tick(time.Duration(*reportInterval) * time.Second) {
		mu.Lock()
		err := sender.SendMetrics(*address, metrics)
		if err != nil {
			log.Printf("Error during send metrics: %s", err)
		} else {
			log.Println("Metrics sent successfully")
		}
		metrics = nil
		mu.Unlock()
	}
}
