package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	custhttp "github.com/AleksandrMac/fileserver/internal/delivery/http"
	"github.com/AleksandrMac/fileserver/internal/repository"
	"github.com/AleksandrMac/fileserver/internal/usecase"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Config
	storagePath := getEnv("STORAGE_PATH", "./storage")
	apiKey := getEnv("API_KEY", "")
	port := getEnv("PORT", "8080")

	if apiKey == "" {
		log.Fatal().Msg("API_KEY is required")
	}

	// Init
	repo := repository.NewFileRepository(storagePath)
	uc := usecase.NewFileUseCase(repo)
	handler := custhttp.NewHandler(uc, apiKey)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(handler.Metrics)

	// Health & Ready
	r.Get("/health", handler.Health)
	r.Get("/reade", handler.Ready)

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	// Upload
	r.Post("/upload", handler.Upload)

	r.Method("GET", "/*", http.HandlerFunc(handler.ServeFileOrMeta))

	// Server
	addr := ":" + port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Запуск сервера в горутине
	go func() {
		log.Info().Str("addr", addr).Str("storage", storagePath).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutdown signal received, initializing graceful shutdown...")

	// Контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// попытка graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
		return
	}

	log.Info().Msg("server exited gracefully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
