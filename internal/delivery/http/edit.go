package http

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func (h Handler) Edit(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Query().Get("file"), "/")
	if filename == "" || strings.Contains(filename, "..") {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Аноним"
	}

	userId := r.URL.Query().Get("userId")
	if userId == "" {
		userId = "9999"
	}

	err := h.editorUC.EditHtml(w, username, userId, filename)
	if err != nil {
		log.Error().Err(err).Msg("failed build editor page")
	}
}
