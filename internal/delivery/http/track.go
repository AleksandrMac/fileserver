package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/rs/zerolog/log"
)

func (x *Handler) Track(w http.ResponseWriter, r *http.Request) {

	// 1. Проверяем JWT из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	if !x.editorUC.VerifyEditorToken(tokenStr) {
		http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
		return
	}

	payload := domain.EditorCallback{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Обрабатываем только статусы 2 (редактирование) и 3 (нужно сохранить)
	if payload.Status == 2 || payload.Status == 3 {
		filename, err := base64.URLEncoding.DecodeString(payload.Key)
		if err != nil {
			log.Error().Err(err).Msg("failed decode key")
			http.Error(w, "failed decode key", http.StatusBadRequest)
			return
		}

		if payload.Url == "" {
			http.Error(w, "missing url tag", http.StatusBadRequest)
			return
		}

		fullFilename, err := x.fileUC.GetFullPath(string(filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 6. Скачиваем обновлённый документ от Document Server
		resp, err := http.Get(payload.Url)
		if err != nil {
			http.Error(w, "Failed to download updated document", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Download failed: %s", resp.Status), http.StatusInternalServerError)
			return
		}

		// 7. Сохраняем поверх существующего файла
		err = x.fileUC.SaveFile(fullFilename, resp.Body)
		if err != nil {
			http.Error(w, "Failed to write document", http.StatusInternalServerError)
			return
		}
	}

	// 8. ОБЯЗАТЕЛЬНО: вернуть {"error": 0}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"error": 0})
}
