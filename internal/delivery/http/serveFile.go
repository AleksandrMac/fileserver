package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	d "github.com/AleksandrMac/fileserver/internal/delivery"
	"github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/AleksandrMac/fileserver/internal/metrics"
	"github.com/AleksandrMac/fileserver/internal/templates"

	"github.com/rs/zerolog/log"
)

func (h *Handler) ServeFileOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
	w.Header().Set("X-API-Param-meta", "For ZIP files: ?meta=true → returns JSON metadata (name, mod_time)")
	w.WriteHeader(http.StatusOK)
}

var funcMap = template.FuncMap{
	"hasSuffix":  strings.HasSuffix,
	"or":         func(a, b bool) bool { return a || b },
	"and":        func(a, b bool) bool { return a && b },
	"formatTime": func(t time.Time) string { return t.Format("15:04 02.01.2006") },
}

func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Path
	resultType := d.ContentType(r.Header.Get("Accept"))
	head := r.Method == http.MethodHead

	fullPath, err := h.fileUC.GetFullPath(relPath)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	info, err := h.fileUC.FileInfo(fullPath)
	if err != nil {
		log.Warn().Err(err).Str("path", fullPath).Msg("failed to read info")
		http.NotFound(w, r)
		return
	}

	if info == nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		files, err := h.fileUC.List(fullPath)
		if err != nil {
			log.Error().Err(err).Str("path", fullPath).Msg("file list get failed")
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		if resultType == d.ApplictionJSON {
			w.Header().Set("Content-Type", string(resultType))
			if head {
				w.WriteHeader(http.StatusOK)
				return
			}
			err = json.NewEncoder(w).Encode(files)
		} else {
			w.Header().Set("Content-Type", string(d.TextHTML))
			if head {
				w.WriteHeader(http.StatusOK)
				return
			}
			err = template.Must(
				template.New("index.html").
					Funcs(funcMap).
					ParseFS(templates.HTML, "html/index.html")).
				Execute(w, map[string]any{
					"Files": files,
					"Dir":   strings.TrimPrefix(relPath, h.urlPrefix),
				})
		}
		if err != nil {
			log.Error().Err(err).
				Str("path", fullPath).
				Str("result-content-type", string(resultType)).
				Msg("failed to encode")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if resultType == d.ApplictionJSON {
		data, err := json.Marshal(domain.FileInfo{
			Name:    info.Name(),
			Path:    relPath,
			IsDir:   false,
			ModTime: info.ModTime(),
		})
		if err != nil {
			log.Error().Err(err).Msg("failed fail info marshal")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.Header().Set("Content-Type", string(d.ApplictionJSON))

		if head {
			w.WriteHeader(http.StatusOK)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			log.Error().Err(err).Msg("failed write to client")
			return
		}
	} else {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		w.Header().Set("Content-Type", string(d.ApplcationOctetStream))

		if head {
			w.WriteHeader(http.StatusOK)
			return
		}

		file, err := h.fileUC.ReadFile(fullPath)
		if err != nil {
			log.Error().Err(err).Str("path", fullPath).Msg("failed read file")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		n, err := io.Copy(w, file)
		if err == nil {
			metrics.BytesDownloaded.Add(float64(n))
		}
	}
}
