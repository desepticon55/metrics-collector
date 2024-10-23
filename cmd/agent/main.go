package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/agent"
	"github.com/desepticon55/metrics-collector/internal/common"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	config := extractConfig()
	flag.Parse()
	logger.Info("Current config:", zap.String("config", config.String()))

	metricCh := make(chan []common.MetricRequestDto, 10)
	var wg sync.WaitGroup

	runtimeProvider := &agent.RuntimeMetricProvider{}
	virtualProvider := &agent.VirtualMetricProvider{}

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigCh
		logger.Info("Shutdown signal received, waiting for graceful shutdown")
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metricCh <- runtimeProvider.GetMetrics()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metricCh <- virtualProvider.GetMetrics()
			}
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
			case <-ctx.Done():
				if len(metrics) > 0 {
					sendRemainingMetrics(sender, metrics, rateLimiter, config, logger)
				}
				return
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
				sendMetrics(sender, metrics, config, logger)
				metrics = nil
			}
		}
	}()

	wg.Wait()
	logger.Info("Graceful shutdown completed")

}

func extractConfig() agent.Config {
	return agent.GetConfig(func(filePath string) (agent.Config, error) {
		var config agent.Config
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return config, fmt.Errorf("could not read config file: %w", err)
		}

		err = json.Unmarshal(fileContent, &config)
		if err != nil {
			return config, fmt.Errorf("could not unmarshal config JSON: %w", err)
		}

		return config, nil
	})
}

func runPprofServer() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6070", nil))
	}()
}

func sendMetrics(sender agent.MetricsSender, metrics []common.MetricRequestDto, config agent.Config, logger *zap.Logger) {
	var err error
	if config.EnabledHTTPS {
		err = sender.SendMetrics(fmt.Sprintf("https://%s/updates/", config.ServerAddress), metrics)
	} else {
		err = sender.SendMetrics(fmt.Sprintf("http://%s/updates/", config.ServerAddress), metrics)
	}
	if err != nil {
		logger.Error("Error during send metrics", zap.Error(err))
	} else {
		logger.Info("Metrics sent successfully")
	}
}

func sendRemainingMetrics(sender agent.MetricsSender, metrics []common.MetricRequestDto, rateLimiter *rate.Limiter, config agent.Config, logger *zap.Logger) {
	logger.Info("Sending remaining metrics before shutdown")
	err := rateLimiter.Wait(context.Background())
	if err != nil {
		logger.Error("Error during rate limit wait", zap.Error(err))
		return
	}
	sendMetrics(sender, metrics, config, logger)
}
