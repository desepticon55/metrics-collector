package agent

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetConfig_Defaults(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Clearenv()
	os.Setenv("CONFIG", "./some/path")

	loadConfig := func(filePath string) (Config, error) {
		return Config{}, nil
	}

	config := GetConfig(loadConfig)

	assert.Equal(t, "localhost:8080", config.ServerAddress)
	assert.Equal(t, 2, config.PollInterval)
	assert.Equal(t, 10, config.ReportInterval)
	assert.Equal(t, "", config.HashKey)
	assert.Equal(t, 1, config.RateLimit)
	assert.False(t, config.EnabledHTTPS)
	assert.Equal(t, "", config.CryptoKey)
}

func TestGetConfig_EnvVars(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Clearenv()
	os.Setenv("CONFIG", "./some/path")

	os.Setenv("ADDRESS", "env-server:9000")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("KEY", "envHashKey")
	os.Setenv("RATE_LIMIT", "3")
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("CRYPTO_KEY", "envCryptoKey")

	loadConfig := func(filePath string) (Config, error) {
		return Config{}, nil
	}

	config := GetConfig(loadConfig)

	assert.Equal(t, "env-server:9000", config.ServerAddress)
	assert.Equal(t, 5, config.PollInterval)
	assert.Equal(t, 15, config.ReportInterval)
	assert.Equal(t, "envHashKey", config.HashKey)
	assert.Equal(t, 3, config.RateLimit)
	assert.True(t, config.EnabledHTTPS)
	assert.Equal(t, "envCryptoKey", config.CryptoKey)
}

func TestGetConfig_FileConfig(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Clearenv()
	os.Setenv("CONFIG", "./some/path")

	fileConfig := Config{
		ServerAddress:  "file-server:8080",
		PollInterval:   2,
		ReportInterval: 10,
		HashKey:        "fileHashKey",
		RateLimit:      1,
		EnabledHTTPS:   true,
		CryptoKey:      "fileCryptoKey",
	}

	loadConfig := func(filePath string) (Config, error) {
		return fileConfig, nil
	}

	config := GetConfig(loadConfig)

	assert.Equal(t, "file-server:8080", config.ServerAddress)
	assert.Equal(t, 2, config.PollInterval)
	assert.Equal(t, 10, config.ReportInterval)
	assert.Equal(t, "fileHashKey", config.HashKey)
	assert.Equal(t, 1, config.RateLimit)
	assert.True(t, config.EnabledHTTPS)
	assert.Equal(t, "fileCryptoKey", config.CryptoKey)
}
