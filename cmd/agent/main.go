package main

import (
	"context"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/agent"
	"github.com/desepticon55/metrics-collector/internal/common"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	runPprofServer()
	logger, err := common.NewLogger()
	if err != nil {
		log.Fatal("Error during initialise logger", err)
	}
	defer logger.Sync()

	config := agent.GetConfig()
	logger.Info("Current config:", zap.String("config", config.String()))

	metricCh := make(chan []common.MetricRequestDto, 10)
	var wg sync.WaitGroup

	runtimeProvider := &agent.RuntimeMetricProvider{}
	virtualProvider := &agent.VirtualMetricProvider{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
		defer ticker.Stop()

		for range time.Tick(time.Duration(config.ReportInterval) * time.Second) {
			metricCh <- runtimeProvider.GetMetrics()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
		defer ticker.Stop()

		for range time.Tick(time.Duration(config.ReportInterval) * time.Second) {
			metricCh <- virtualProvider.GetMetrics()
		}
	}()

	sender := agent.New(config)
	rateLimiter := rate.NewLimiter(rate.Every(1*time.Second), config.RateLimit)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
		defer ticker.Stop()

		var metrics []common.MetricRequestDto

		for {
			select {
			case newMetrics := <-metricCh:
				metrics = append(metrics, newMetrics...)
			case <-ticker.C:
				if len(metrics) == 0 {
					continue
				}
				err := rateLimiter.Wait(context.Background())
				if err != nil {
					logger.Error("Error during rate limit wait", zap.Error(err))
					continue
				}
				err = sender.SendMetrics(fmt.Sprintf("http://%s/updates/", config.ServerAddress), metrics)
				if err != nil {
					logger.Error("Error during send metrics", zap.Error(err))
				} else {
					logger.Info("Metrics sent successfully")
					metrics = nil
				}
			}
		}
	}()
	wg.Wait()

}

func runPprofServer() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6070", nil))
	}()
}
