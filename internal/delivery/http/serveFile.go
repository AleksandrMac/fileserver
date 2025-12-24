package http

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AleksandrMac/fileserver/internal/metrics"
	"github.com/rs/zerolog/log"
)

func (h *Handler) ServeFileGet(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Path
	relPath = filepath.Clean("/" + relPath)
	if relPath == "/" || strings.Contains(relPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath, err := h.usecase.GetFilePath(relPath[1:])
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

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
		if err := json.NewEncoder(w).Encode(files); err != nil {
			log.Error().Err(err).Msg("failed to encode archive metadata")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
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
