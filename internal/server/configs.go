package server

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
}

func GetConfig() Config {
	defaultAddress := "localhost:8080"
	if envAddr, exists := os.LookupEnv("ADDRESS"); exists {
		defaultAddress = envAddr
	}
	address := flag.String("a", defaultAddress, "Server address")
	flag.Parse()
	return Config{
		ServerAddress: *address,
	}
}
