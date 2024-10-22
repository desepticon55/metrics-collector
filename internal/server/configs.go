package server

import (
	"flag"
	"fmt"
	"log"
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
}

type ConfigLoaderFunc func(filePath string) (Config, error)

func (c Config) String() string {
	return fmt.Sprintf("\nServerAddress: %s\nDatabaseConnString: %s\nStoreInterval: %d\nFileStoragePath: %s\nHashKey: %s\nRestore: %t\nEnabledHttps: %t\nCryptoKey: %s",
		c.ServerAddress, c.DatabaseConnString, c.StoreInterval, c.FileStoragePath, c.HashKey, c.Restore, c.EnabledHTTPS, c.CryptoKey)
}

func ParseConfig(loadConfig ConfigLoaderFunc) Config {
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
	fileStoragePath := getStringValue(os.Getenv("FILE_STORAGE_PATH"), *flag.String("f", "", "File storage path"), fileConfig.FileStoragePath, "C:\\opt\\metrics-db.json")
	databaseConnString := getStringValue(os.Getenv("DATABASE_DSN"), *flag.String("d", "", "Database connection string"), fileConfig.DatabaseConnString, "postgres://postgres:postgres@localhost:5432/postgres")
	hashKey := getStringValue(os.Getenv("KEY"), *flag.String("k", "", "Hash key"), fileConfig.HashKey, "")
	cryptoKey := getStringValue(os.Getenv("CRYPTO_KEY"), *flag.String("crypto-key", "", "Crypto key"), fileConfig.CryptoKey, "")
	restore := getBooleanValue(os.Getenv("RESTORE"), *flag.Bool("r", false, "Load data from file or not"), fileConfig.Restore)
	enableHTTPS := getBooleanValue(os.Getenv("ENABLE_HTTPS"), *flag.Bool("s", false, "Load data from file or not"), fileConfig.EnabledHTTPS)
	storeInterval := getIntValue(os.Getenv("STORE_INTERVAL"), *flag.Int("i", 5, "Store interval (sec.)"), fileConfig.StoreInterval)

	return Config{
		ServerAddress:      address,
		StoreInterval:      storeInterval,
		FileStoragePath:    fileStoragePath,
		Restore:            restore,
		DatabaseConnString: databaseConnString,
		HashKey:            hashKey,
		EnabledHTTPS:       enableHTTPS,
		CryptoKey:          cryptoKey,
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
