package main

import (
	"io"
	"net/http"

	"github.com/kbannyi/gophermart/internal/config"
)

func main() {
	cfg := config.ParseConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello, World!")
	})
	http.ListenAndServe(cfg.RunAddr, mux)
}
