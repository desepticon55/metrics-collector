package server

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestParseConfig_ShouldReturnDefaultValues(t *testing.T) {
	os.Clearenv()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	config := CreateConfig(nil, func(filePath string) (Config, error) {
		return Config{}, nil
	})

	assert.Equal(t, "localhost:8080", config.ServerAddress)
	assert.Equal(t, "C:\\opt\\metrics-db.json", config.FileStoragePath)
	assert.Equal(t, "postgres://postgres:postgres@localhost:5432/postgres", config.DatabaseConnString)
	assert.Equal(t, 5, config.StoreInterval)
	assert.False(t, config.Restore)
	assert.False(t, config.EnabledHTTPS)
	assert.Empty(t, config.HashKey)
	assert.Empty(t, config.CryptoKey)
}

func TestParseConfig_ShouldReturnEnvOverrides(t *testing.T) {
	os.Setenv("ADDRESS", "127.0.0.1:9000")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/metrics.json")
	os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/db")
	os.Setenv("STORE_INTERVAL", "10")
	os.Setenv("RESTORE", "true")
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("KEY", "testhash")
	os.Setenv("CRYPTO_KEY", "secretkey")
	os.Setenv("TRUSTED_SUBNET", "172.18.208.1/32")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	config := CreateConfig(nil, func(filePath string) (Config, error) {
		return Config{}, nil
	})

	assert.Equal(t, "127.0.0.1:9000", config.ServerAddress)
	assert.Equal(t, "/tmp/metrics.json", config.FileStoragePath)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", config.DatabaseConnString)
	assert.Equal(t, 10, config.StoreInterval)
	assert.True(t, config.Restore)
	assert.True(t, config.EnabledHTTPS)
	assert.Equal(t, "testhash", config.HashKey)
	assert.Equal(t, "secretkey", config.CryptoKey)
	assert.Equal(t, "172.18.208.1/32", config.TrustedSubnet)
}

func TestParseConfig_ShouldReturnFileOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("CONFIG", "./some/path")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	config := CreateConfig(nil, func(filePath string) (Config, error) {
		return Config{
			ServerAddress:      "192.168.1.1:8000",
			FileStoragePath:    "/data/metrics.json",
			DatabaseConnString: "postgres://fileuser:filepass@db:5432/filedb",
			StoreInterval:      5,
			Restore:            true,
			EnabledHTTPS:       true,
			HashKey:            "filehash",
			CryptoKey:          "filecrypto",
			TrustedSubnet:      "172.18.208.1/32",
		}, nil
	})

	assert.Equal(t, "192.168.1.1:8000", config.ServerAddress)
	assert.Equal(t, "/data/metrics.json", config.FileStoragePath)
	assert.Equal(t, "postgres://fileuser:filepass@db:5432/filedb", config.DatabaseConnString)
	assert.Equal(t, 5, config.StoreInterval)
	assert.True(t, config.Restore)
	assert.True(t, config.EnabledHTTPS)
	assert.Equal(t, "filehash", config.HashKey)
	assert.Equal(t, "filecrypto", config.CryptoKey)
	assert.Equal(t, "172.18.208.1/32", config.TrustedSubnet)
}
