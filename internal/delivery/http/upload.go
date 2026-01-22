package http

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/AleksandrMac/fileserver/internal/metrics"
	"github.com/rs/zerolog/log"
)

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Missing 'filename' query param", http.StatusBadRequest)
		return
	}

	relPath := r.URL.Path
	fullPath, err := h.fileUC.GetFullPath(relPath)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	storeInfo, err := h.fileUC.FileInfo(fullPath)
	if storeInfo == nil {
		if err != nil {
			log.Warn().Err(err).Str("path", fullPath).Msg("failed to read info")
		}
		http.NotFound(w, r)
		return
	}

	if !storeInfo.IsDir() {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename = filepath.Clean("/" + filename)
	if strings.Contains(filename, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullFileName := filepath.Join(fullPath, filename)

	oldFileInfo, err := h.fileUC.FileInfo(fullFileName)
	if err != nil {
		log.Error().Err(err).Str("path", fullPath).Msg("get info failed")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Save
	if err := h.fileUC.SaveFile(fullFileName, file); err != nil {
		log.Error().Err(err).Str("path", fullPath).Msg("upload failed")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	newSize, _ := h.fileUC.GetFileSize(fullFileName)

	var delta int64

	if oldFileInfo != nil {
		delta = newSize - oldFileInfo.Size()
	} else {
		delta = newSize
	}

	h.storageSize += delta
	metrics.TotalStorageSize.Set(float64(h.storageSize))
	metrics.BytesUploaded.Add(float64(newSize))

	w.WriteHeader(http.StatusCreated)
	log.Info().Str("path", filename).Int64("size", newSize).Msg("file uploaded")
}

// UploadOptions обрабатывает CORS preflight для /upload
func (h *Handler) UploadOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "POST, OPTIONS, HEAD")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// Описание параметров через кастомные заголовки (для разработчиков)
	w.Header().Set("X-API-Param-path", "Required. Relative file path for upload (e.g., ?path=docs/report.pdf). Must not contain '..' or absolute paths.")
	w.Header().Set("X-API-Header-X-API-Key", "Required for POST. API key for authorization.")

	w.WriteHeader(http.StatusOK)
}

// UploadHead проверяет доступность эндпоинта /upload
func (h *Handler) UploadHead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "POST, OPTIONS, HEAD")
	w.Header().Set("X-API-Param-path", "Required. Relative file path for upload (e.g., ?path=docs/report.pdf). Must not contain '..' or absolute paths.")
	w.Header().Set("X-API-Header-X-API-Key", "Required for POST. API key for authorization.")
	w.WriteHeader(http.StatusOK)
}
