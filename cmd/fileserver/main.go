package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	custhttp "github.com/AleksandrMac/fileserver/internal/delivery/http"
	"github.com/AleksandrMac/fileserver/internal/repository"
	"github.com/AleksandrMac/fileserver/internal/usecase"
	editor_usecase "github.com/AleksandrMac/fileserver/internal/usecase/editor"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Эти переменные заполняются при сборке через -ldflags
var (
	version   string // например: "v1.2.0" или "dev"
	commit    string // хеш коммита
	buildTime string // ISO8601 время
)

func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// default value
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	// Config
	storagePath := getEnv("STORAGE_PATH", "./storage")
	apiKey := getEnv("API_KEY", "")
	port := getEnv("PORT", "8080")
	hostname = getEnv("HOST", hostname)
	jwtSecret := getEnv("DOCUMENT_SERVER_SECRET", "")
	docServerUrl := getEnv("DOCUMENT_SERVER_URL", "")
	docServerUrlInternal := getEnv("DOCUMENT_SERVER_URL_INTERNAL", "")
	storageUrlPath := storagePathUrl()
	if err := os.MkdirAll(filepath.Join(storagePath, storageUrlPath), 0755); err != nil {
		log.Fatal().Msg("can't make storage")
	}

	if apiKey == "" {
		log.Fatal().Msg("API_KEY is required")
	}
	if jwtSecret == "" {
		log.Fatal().Msg("DOCUMENT_SERVER_SECRET is required")
	}
	if docServerUrl == "" {
		log.Fatal().Msg("DOCUMENT_SERVER_URL is required")
	}

	// Init
	repo := repository.NewFileRepository(storagePath)
	fileUC := usecase.NewFileUseCase(repo)
	infoUC := usecase.NewInfoService(version, commit, buildTime, port, repo)
	editorUC := editor_usecase.NewEditorUsecase(jwtSecret, docServerUrl, docServerUrlInternal, fmt.Sprintf("http://%s:%s", hostname, port))
	trackUC := usecase.NewTrackUC(repo, docServerUrl, docServerUrlInternal)
	handler := custhttp.NewHandler(fileUC, infoUC, editorUC, trackUC, apiKey, storageUrlPath)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(handler.Metrics)

	// Health & Ready
	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)
	r.Get("/info", handler.Info)

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	r.Get(storageUrlPath+"*", handler.ServeFile)
	r.Post(storageUrlPath+"*", handler.Auth(http.HandlerFunc(handler.Upload)).ServeHTTP)
	r.Head(storageUrlPath+"*", handler.ServeFile)
	r.Options(storageUrlPath+"*", handler.ServeFileOptions)

	r.Get("/edit", handler.Edit)
	r.Post("/track", handler.Track)

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

func storagePathUrl() string {
	path := getEnv("STORAGE_PATH_URL", "/")
	path, _ = url.JoinPath("/", path)
	return path
}
