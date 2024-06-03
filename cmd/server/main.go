package main

import (
	"github.com/desepticon55/metrics-collector/internal/server"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/update/", server.NewWriteMetricHandler(server.NewMemStorage()))
	mux.Handle("/find/", server.NewReadMetricHandler(server.NewMemStorage()))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
