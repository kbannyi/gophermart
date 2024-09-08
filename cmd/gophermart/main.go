package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kbannyi/gophermart/internal/config"
	"github.com/kbannyi/gophermart/internal/handler"
	"github.com/kbannyi/gophermart/internal/logger"
)

func main() {
	logger.Initialize()
	cfg := config.ParseConfig()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", handler.NewHealthHandler())

	logger.Log.Info("Starting server...")
	http.ListenAndServe(cfg.RunAddr, r)
}
