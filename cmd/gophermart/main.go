package main

import (
	"io"
	"net/http"

	"github.com/kbannyi/gophermart/internal/config"
	"github.com/kbannyi/gophermart/internal/logger"
)

func main() {
	logger.Initialize()
	cfg := config.ParseConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello, World!")
	})
	logger.Log.Info("Starting server...")
	http.ListenAndServe(cfg.RunAddr, mux)
}
