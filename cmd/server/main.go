package main

import (
	"github.com/desepticon55/metrics-collector/internal/server"
	"net/http"
)

func main() {
	storage := server.NewMemStorage()

	mux := http.NewServeMux()
	mux.Handle("/update/", server.NewWriteMetricHandler(storage))
	mux.Handle("/find/", server.NewReadMetricHandler(storage))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
