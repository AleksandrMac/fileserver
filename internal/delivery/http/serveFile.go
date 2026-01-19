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

func (h *Handler) ServeFileOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
	w.Header().Set("X-API-Param-meta", "For ZIP files: ?meta=true → returns JSON metadata (name, mod_time)")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ServeFileGet(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Path
	relPath = filepath.Clean("/" + relPath)
	if relPath == "/" || strings.Contains(relPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath, err := h.fileUC.GetFilePath(relPath[1:])
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	exists, _, err := h.fileUC.FileExists(fullPath)
	if err != nil || !exists {
		http.NotFound(w, r)
		return
	}

	isZip := strings.HasSuffix(strings.ToLower(fullPath), ".zip")
	head := r.Method == http.MethodHead
	meta := r.URL.Query().Get("meta") == "true"

	if meta && isZip {
		w.Header().Set("Content-Type", "application/json")
		if head {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		files, err := h.fileUC.ListZipContents(fullPath)
		if err != nil {
			log.Error().Err(err).Str("path", fullPath).Msg("failed to read zip")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}	

		if err := json.NewEncoder(w).Encode(files); err != nil {
			log.Error().Err(err).Msg("failed to encode archive metadata")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	size, _ := h.fileUC.GetFileSize(fullPath)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Content-Type", "application/octet-stream")

	if head {
		w.WriteHeader(http.StatusOK)
		return
	}

	file, err := h.fileUC.ReadFile(fullPath)
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
