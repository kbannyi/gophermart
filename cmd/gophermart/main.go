package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	r.Route("/health", func(r chi.Router) {
		h := handler.NewHealthHandler()
		r.Get("/ping", h.Ping)
	})
	r.Route("/api/user", func(r chi.Router) {
		h := handler.NewAuthHandler()
		r.Post("/register", h.RegisterUser)
		r.Post("/login", h.LoginUser)
	})

	run(cfg, r)
}

func run(cfg config.Config, h http.Handler) {
	server := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: h}
	go func() {
		logger.Log.Info("Starting server...")
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("HTTP server error: ", "err", err)
		}
		logger.Log.Info("Server stopped")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("HTTP server shutdown error", "err", err)
	}
}
