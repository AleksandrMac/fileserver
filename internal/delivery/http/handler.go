package http

import (
	"net/http"
	"strconv"
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

}

func (h *Handler) ServeFileOrMeta(w http.ResponseWriter, r *http.Request) {

}

// middleware

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
