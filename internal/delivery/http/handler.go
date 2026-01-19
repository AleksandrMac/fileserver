package http

import (
	"net/http"

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

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
