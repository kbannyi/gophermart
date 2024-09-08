package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kbannyi/gophermart/internal/config"
	"github.com/kbannyi/gophermart/internal/handler"
	"github.com/kbannyi/gophermart/internal/logger"
)

func main() {
	logger.Initialize()
	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if err := migrateDB(cfg); err != nil {
		logger.Log.Error(err.Error())
		return
	}

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

func migrateDB(cfg config.Config) error {
	logger.Log.Info("Applying DB migrations...")
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %w", err)
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("Unable to create migration driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("Unable to create migrator instance: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Unable to apply migrations: %w", err)
	}
	logger.Log.Info("DB migrations applied")

	return nil
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
