package http

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/AleksandrMac/fileserver/internal/metrics"
	"github.com/rs/zerolog/log"
)

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

	fullPath, err := h.usecase.GetFilePath(relPath[1:])
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

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
