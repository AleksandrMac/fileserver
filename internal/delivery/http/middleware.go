package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AleksandrMac/fileserver/internal/metrics"
	"github.com/rs/zerolog/log"
)

func (h *Handler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkList := []func() bool{
			func() bool { return r.Header.Get("X-API-Key") == h.apiKey },
			func() bool {
				t := r.Header.Get("Authorization")
				return strings.HasPrefix(t, "Bearer ") &&
					h.editorUC.VerifyEditorToken(strings.TrimPrefix(t, "Bearer "))
			},
		}

		for _, f := range checkList {
			if f() {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

func (h *Handler) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		metrics.RequesCount.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(ww.statusCode),
		).Inc()

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration", duration).
			Msg("request completed")
	})
}
