package server

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress      string `json:"address"`
	FileStoragePath    string `json:"store_file"`
	StoreInterval      int    `json:"store_interval"`
	Restore            bool   `json:"restore"`
	DatabaseConnString string `json:"database_dsn"`
	HashKey            string `json:"hash_key"`
	EnabledHTTPS       bool   `json:"enabled_https"`
	CryptoKey          string `json:"crypto_key"`
	TrustedSubnet      string `json:"trusted_subnet"`
	EnabledGRPC        bool   `json:"enabled_grpc"`
}

func (c Config) String() string {
	return fmt.Sprintf("\nServerAddress: %s\nDatabaseConnString: %s\nStoreInterval: %d\nFileStoragePath: %s\nHashKey: %s\nRestore: %t\nEnabledHttps: %t\nCryptoKey: %s",
		c.ServerAddress, c.DatabaseConnString, c.StoreInterval, c.FileStoragePath, c.HashKey, c.Restore, c.EnabledHTTPS, c.CryptoKey)
}

func CreateConfig(logger *zap.Logger, loadConfig func(filePath string) (Config, error)) Config {
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
			logger.Error("Failed to load config from file", zap.Error(err))
		}
	}

	address := getStringValue(os.Getenv("ADDRESS"), *flag.String("a", "", "Server address"), fileConfig.ServerAddress, "localhost:8080")
	fileStoragePath := getStringValue(os.Getenv("FILE_STORAGE_PATH"), *flag.String("f", "", "File storage path"), fileConfig.FileStoragePath, "C:\\opt\\metrics-db.json")
	databaseConnString := getStringValue(os.Getenv("DATABASE_DSN"), *flag.String("d", "", "Database connection string"), fileConfig.DatabaseConnString, "postgres://postgres:postgres@localhost:5432/postgres")
	hashKey := getStringValue(os.Getenv("KEY"), *flag.String("k", "", "Hash key"), fileConfig.HashKey, "")
	cryptoKey := getStringValue(os.Getenv("CRYPTO_KEY"), *flag.String("crypto-key", "", "Crypto key"), fileConfig.CryptoKey, "")
	restore := getBooleanValue(os.Getenv("RESTORE"), *flag.Bool("r", false, "Load data from file or not"), fileConfig.Restore)
	enableHTTPS := getBooleanValue(os.Getenv("ENABLE_HTTPS"), *flag.Bool("s", false, "Load data from file or not"), fileConfig.EnabledHTTPS)
	storeInterval := getIntValue(os.Getenv("STORE_INTERVAL"), *flag.Int("i", 5, "Store interval (sec.)"), fileConfig.StoreInterval)
	trustedSubnet := getStringValue(os.Getenv("TRUSTED_SUBNET"), *flag.String("t", "", "Trusted subnet in CIDR format"), fileConfig.TrustedSubnet, "")
	enableGRPC := getBooleanValue(os.Getenv("ENABLE_GRPC"), *flag.Bool("g", false, "Enabled GRPC or not"), fileConfig.EnabledHTTPS)

	return Config{
		ServerAddress:      address,
		StoreInterval:      storeInterval,
		FileStoragePath:    fileStoragePath,
		Restore:            restore,
		DatabaseConnString: databaseConnString,
		HashKey:            hashKey,
		EnabledHTTPS:       enableHTTPS,
		CryptoKey:          cryptoKey,
		TrustedSubnet:      trustedSubnet,
		EnabledGRPC:        enableGRPC,
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
