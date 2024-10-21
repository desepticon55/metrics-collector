package server

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress      string
	FileStoragePath    string
	StoreInterval      int
	Restore            bool
	DatabaseConnString string
	HashKey            string
	EnabledHTTPS       bool
	CryptoKey          string
}

func (c Config) String() string {
	return fmt.Sprintf("\nServerAddress: %s\nDatabaseConnString: %s\nStoreInterval: %d\nFileStoragePath: %s\nHashKey: %s\nRestore: %t\nEnabledHttps: %t\nCryptoKey: %s",
		c.ServerAddress, c.DatabaseConnString, c.StoreInterval, c.FileStoragePath, c.HashKey, c.Restore, c.EnabledHTTPS, c.CryptoKey)
}

func ParseConfig() Config {
	defaultAddress := "localhost:8080"
	if envAddr, exists := os.LookupEnv("ADDRESS"); exists {
		defaultAddress = envAddr
	}
	address := flag.String("a", defaultAddress, "Server address")

	defaultFileStoragePath := "C:\\opt\\metrics-db.json"
	if envFileStoragePath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		defaultFileStoragePath = envFileStoragePath
	}
	fileStoragePath := flag.String("f", defaultFileStoragePath, "File storage path")

	defaultStoreInterval := 5
	if envStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		if parsedStoreInterval, err := strconv.Atoi(envStoreInterval); err == nil {
			defaultStoreInterval = parsedStoreInterval
		}
	}
	storeInterval := flag.Int("i", defaultStoreInterval, "Store interval (sec.)")

	defaultRestore := true
	if envRestore, exists := os.LookupEnv("RESTORE"); exists {
		if parsedRestore, err := strconv.ParseBool(envRestore); err == nil {
			defaultRestore = parsedRestore
		}
	}
	restore := flag.Bool("r", defaultRestore, "Load data from file or not")

	defaultEnableHTTPS := false
	if envEnableHTTPS, exists := os.LookupEnv("ENABLE_HTTPS"); exists {
		if parsedEnableHTTPS, err := strconv.ParseBool(envEnableHTTPS); err == nil {
			defaultEnableHTTPS = parsedEnableHTTPS
		}
	}
	enableHTTPS := flag.Bool("s", defaultEnableHTTPS, "Enable HTTPS or not")

	defaultDatabaseConnString := "postgres://postgres:postgres@localhost:5432/postgres"
	if envDatabaseConnString, exists := os.LookupEnv("DATABASE_DSN"); exists {
		defaultDatabaseConnString = envDatabaseConnString
	}
	databaseConnString := flag.String("d", defaultDatabaseConnString, "Server address")

	defaultHashKey := ""
	if hashKey, exists := os.LookupEnv("KEY"); exists {
		defaultHashKey = hashKey
	}
	hashKey := flag.String("k", defaultHashKey, "Hash key")

	defaultCryptoKey := ""
	if envCryptoKey, exists := os.LookupEnv("CRYPTO_KEY"); exists {
		defaultCryptoKey = envCryptoKey
	}
	cryptoKey := flag.String("crypto-key", defaultCryptoKey, "Crypto key")

	flag.Parse()
	return Config{
		ServerAddress:      *address,
		StoreInterval:      *storeInterval,
		FileStoragePath:    *fileStoragePath,
		Restore:            *restore,
		DatabaseConnString: *databaseConnString,
		HashKey:            *hashKey,
		EnabledHTTPS:       *enableHTTPS,
		CryptoKey:          *cryptoKey,
	}
}
