package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	chi_middleware "github.com/go-chi/chi/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/gophermart/internal/config"
	"github.com/kbannyi/gophermart/internal/handler"
	"github.com/kbannyi/gophermart/internal/logger"
	"github.com/kbannyi/gophermart/internal/middleware"
	"github.com/kbannyi/gophermart/internal/repository"
	"github.com/kbannyi/gophermart/internal/service"
)

func main() {
	logger.Initialize()
	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("Unable to connect to database: %v", err))
		return
	}
	defer db.Close()
	if err := migrateDB(db); err != nil {
		logger.Log.Error(err.Error())
		return
	}
	dbx := sqlx.NewDb(db, "pgx")
	userRepository := repository.NewUserRepository(dbx)
	authService := service.NewAuthService(userRepository)
	jobctx, jobcancel := context.WithCancel(context.Background())
	var jobwg sync.WaitGroup

	r := chi.NewRouter()
	r.Use(chi_middleware.Logger)
	r.Use(middleware.AuthExtractor)
	r.Route("/health", func(r chi.Router) {
		h := handler.NewHealthHandler()
		r.Get("/ping", h.Ping)
	})
	r.Route("/api/user", func(r chi.Router) {
		authHandler := handler.NewAuthHandler(authService)
		r.Post("/register", authHandler.RegisterUser)
		r.Post("/login", authHandler.LoginUser)

		r = r.With(middleware.AuthGuard)
		orderRepo := repository.NewOrderRepository(dbx)
		worker := service.NewOrderFetcher(orderRepo, jobctx, &jobwg, cfg)
		orderService := service.NewOrderService(orderRepo, worker)
		orderHandler := handler.NewOrderHandler(orderService)
		worker.Run()
		r.Post("/orders", orderHandler.SaveOrder)
		r.With(chi_middleware.Compress(5)).Get("/orders", orderHandler.GetOrders)

		withdrawalRepository := repository.NewWithdrawalRepository(dbx)
		withdrawalService := service.NewWithdrawalService(withdrawalRepository)
		withdrawalHandler := handler.NewWithdrawalHandler(withdrawalService)
		r.Get("/balance", withdrawalHandler.GetBalance)
		r.Post("/balance/withdraw", withdrawalHandler.Withdraw)
		r.With(chi_middleware.Compress(5)).Get("/withdrawals", withdrawalHandler.GetWithdrawals)
	})

	run(cfg, r)
	jobcancel()
	jobwg.Wait()
	logger.Log.Info("Exiting gracefully")
}

func migrateDB(db *sql.DB) error {
	logger.Log.Info("Applying DB migrations...")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("unable to create migration driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("unable to create migrator instance: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("unable to apply migrations: %w", err)
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
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigChan

	logger.Log.Info("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("HTTP server shutdown error", "err", err)
	}
}
