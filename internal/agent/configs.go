package agent

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Agent configuration
type Config struct {
	ServerAddress  string
	PollInterval   int
	ReportInterval int
	HashKey        string
	RateLimit      int
}

func (c Config) String() string {
	return fmt.Sprintf("\nServerAddress: %s\nPollInterval: %d\nReportInterval: %d\nHashKey: %s\n",
		c.ServerAddress, c.PollInterval, c.ReportInterval, c.HashKey)
}

func GetConfig() Config {
	defaultAddress := "localhost:8080"
	if envAddr, exists := os.LookupEnv("ADDRESS"); exists {
		defaultAddress = envAddr
	}
	address := flag.String("a", defaultAddress, "Server address")

	defaultPollInterval := 2
	if envPollInterval, exists := os.LookupEnv("POLL_INTERVAL"); exists {
		if parsedPollInterval, err := strconv.Atoi(envPollInterval); err == nil {
			defaultPollInterval = parsedPollInterval
		}
	}
	pollInterval := flag.Int("p", defaultPollInterval, "Poll interval (sec.)")

	defaultReportInterval := 10
	if envReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); exists {
		if parsedReportInterval, err := strconv.Atoi(envReportInterval); err == nil {
			defaultReportInterval = parsedReportInterval
		}
	}
	reportInterval := flag.Int("r", defaultReportInterval, "Report interval (sec.)")

	defaultHashKey := ""
	if hashKey, exists := os.LookupEnv("KEY"); exists {
		defaultHashKey = hashKey
	}
	hashKey := flag.String("k", defaultHashKey, "Hash key")

	defaultRateLimit := 1
	if rateLimit, exists := os.LookupEnv("RATE_LIMIT"); exists {
		if parsedRateLimit, err := strconv.Atoi(rateLimit); err == nil {
			defaultRateLimit = parsedRateLimit
		}
	}
	rateLimit := flag.Int("l", defaultRateLimit, "Rate limit")

	flag.Parse()
	return Config{
		ServerAddress:  *address,
		PollInterval:   *pollInterval,
		ReportInterval: *reportInterval,
		HashKey:        *hashKey,
		RateLimit:      *rateLimit,
	}
}
