package agent

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Agent configuration
type Config struct {
	ServerAddress  string `json:"address"`
	PollInterval   int    `json:"poll_interval"`
	ReportInterval int    `json:"report_interval"`
	HashKey        string `json:"hash_key"`
	RateLimit      int    `json:"rate_limit"`
	EnabledHTTPS   bool   `json:"enabled_https"`
	CryptoKey      string `json:"crypto_key"`
}

func (c Config) String() string {
	return fmt.Sprintf("\nServerAddress: %s\nPollInterval: %d\nReportInterval: %d\nHashKey: %s\n",
		c.ServerAddress, c.PollInterval, c.ReportInterval, c.HashKey)
}

func GetConfig(loadConfig func(filePath string) (Config, error)) Config {
	defaultConfigPath := ""
	if envConfigPath, exists := os.LookupEnv("CONFIG"); exists {
		defaultConfigPath = envConfigPath
	}
	configPath := flag.String("config", defaultConfigPath, "Path to config file")

	var fileConfig Config
	if *configPath != "" {
		var err error
		fileConfig, err = loadConfig(*configPath)
		if err != nil {
			log.Printf("Failed to load config from file: %v", err)
		}
	}

	address := getStringValue(os.Getenv("ADDRESS"), *flag.String("a", "", "Server address"), fileConfig.ServerAddress, "localhost:8080")
	pollInterval := getIntValue(os.Getenv("POLL_INTERVAL"), *flag.Int("p", 2, "Poll interval (sec.)"), fileConfig.PollInterval)
	reportInterval := getIntValue(os.Getenv("REPORT_INTERVAL"), *flag.Int("r", 10, "Report interval (sec.)"), fileConfig.ReportInterval)
	hashKey := getStringValue(os.Getenv("KEY"), *flag.String("k", "", "Hash key"), fileConfig.HashKey, "")
	rateLimit := getIntValue(os.Getenv("RATE_LIMIT"), *flag.Int("l", 1, "Rate limit"), fileConfig.RateLimit)
	enableHTTPS := getBooleanValue(os.Getenv("ENABLE_HTTPS"), *flag.Bool("s", false, "Load data from file or not"), fileConfig.EnabledHTTPS)
	cryptoKey := getStringValue(os.Getenv("CRYPTO_KEY"), *flag.String("crypto-key", "", "Crypto key"), fileConfig.CryptoKey, "")

	return Config{
		ServerAddress:  address,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		HashKey:        hashKey,
		RateLimit:      rateLimit,
		EnabledHTTPS:   enableHTTPS,
		CryptoKey:      cryptoKey,
	}
}

func getStringValue(env string, flagValue string, fileValue string, defaultValue string) string {
	if env != "" {
		return env
	}
	if flagValue != "" {
		return flagValue
	}
	if fileValue != "" {
		return fileValue
	}
	return defaultValue
}

func getBooleanValue(envVar string, flagValue bool, fileValue bool) bool {
	if envVar != "" {
		if parsed, err := strconv.ParseBool(envVar); err == nil {
			return parsed
		}
	}
	if flagValue {
		return true
	}
	return fileValue
}

func getIntValue(envVar string, flagValue int, fileValue int) int {
	if envVar != "" {
		if parsed, err := strconv.Atoi(envVar); err == nil {
			return parsed
		}
	}
	if flagValue != 0 {
		return flagValue
	}
	return fileValue
}
