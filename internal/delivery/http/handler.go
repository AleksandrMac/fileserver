package http

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/AleksandrMac/fileserver/internal/interfaces"
	"github.com/AleksandrMac/fileserver/internal/metrics"
)

type Handler struct {
	fileUC        interfaces.FileUsecase
	infoServiceUC interfaces.InfoServiceInterface
	editorUC      interfaces.EditorUsecase
	apiKey        string
	storageSize   int64
}

func NewHandler(usecase interfaces.FileUsecase, infoService interfaces.InfoServiceInterface, editor interfaces.EditorUsecase, apiKey string) *Handler {
	storage, err := usecase.GetStorageInfo()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to calculate initial storage size")
	}

	log.Info().Int64("bytes", storage.TotalSize).Msg("initial storage size calculated")
	metrics.TotalStorageSize.Set(float64(storage.TotalSize))

	return &Handler{
		fileUC:        usecase,
		infoServiceUC: infoService,
		editorUC:      editor,
		apiKey:        apiKey,
		storageSize:   storage.TotalSize,
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

func (x *Handler) Info(w http.ResponseWriter, r *http.Request) {
	info := x.infoServiceUC.GetInfo()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Warn().Err(err).Msg("failed to encode info response")
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
