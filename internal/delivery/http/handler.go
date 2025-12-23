package http

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/AleksandrMac/fileserver/internal/interfaces"
	"github.com/AleksandrMac/fileserver/internal/metrics"
)

type Handler struct {
	usecase     interfaces.FileUsecase
	apiKey      string
	storageSize int64
}

func NewHandler(usecase interfaces.FileUsecase, apiKey string) *Handler {
	size, err := usecase.GetStorageSize()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to calculate initial storage size")
	}

	log.Info().Int64("bytes", size).Msg("initial storage size calculated")
	metrics.TotalStorageSize.Set(float64(size))

	return &Handler{
		usecase:     usecase,
		apiKey:      apiKey,
		storageSize: size,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	// TODO: добавить проверку доступности репозитория

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		http.Error(w, "Missing 'path' query param", http.StatusBadRequest)
		return
	}

	relPath = filepath.Clean("/" + relPath)
	if strings.Contains(relPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := h.usecase.GetFilePath(relPath[1:])
	exists, oldSize, _ := h.usecase.FileExists(fullPath)

	// Save
	if err := h.usecase.SaveFile(fullPath, file); err != nil {
		log.Error().Err(err).Str("path", fullPath).Msg("upload failed")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	newSize, _ := h.usecase.GetFileSize(fullPath)

	var delta int64

	if exists {
		delta = newSize - oldSize
	} else {
		delta = newSize
	}

	h.storageSize += delta
	metrics.TotalStorageSize.Set(float64(h.storageSize))
	metrics.BytesUploaded.Add(float64(newSize))

	w.WriteHeader(http.StatusCreated)
	log.Info().Str("path", relPath).Int64("size", newSize).Msg("file uploaded")
}

func (h *Handler) ServeFileGet(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Path
	relPath = filepath.Clean("/" + relPath)
	if relPath == "/" || strings.Contains(relPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := h.usecase.GetFilePath(relPath[1:])
	exists, _, err := h.usecase.FileExists(fullPath)
	if err != nil || !exists {
		http.NotFound(w, r)
		return
	}

	isZip := strings.HasSuffix(strings.ToLower(fullPath), ".zip")
	head := r.Method == http.MethodHead

	if head && isZip {
		files, err := h.usecase.ListZipContents(fullPath)
		if err != nil {
			log.Error().Err(err).Str("path", fullPath).Msg("failed to read zip")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
		w.WriteHeader(http.StatusOK)
		return
	}

	size, _ := h.usecase.GetFileSize(fullPath)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Content-Type", "application/octet-stream")

	if head {
		w.WriteHeader(http.StatusOK)
		return
	}

	file, err := h.usecase.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	n, err := io.Copy(w, file)
	if err == nil {
		metrics.BytesDownloaded.Add(float64(n))
	}
}

func (h *Handler) ServeFileOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.WriteHeader(http.StatusOK)
}

// middleware

func (h *Handler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key != h.apiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		labels := []string{
			"method", r.Method,
			"path", r.URL.Path,
			"status", strconv.Itoa(ww.statusCode),
		}
		metrics.RequesCount.WithLabelValues(labels...).Inc()

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration", duration).
			Msg("request completed")
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
