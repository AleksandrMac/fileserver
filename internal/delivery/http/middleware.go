package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/AleksandrMac/fileserver/internal/metrics"
	"github.com/rs/zerolog/log"
)

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
