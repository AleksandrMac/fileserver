package editor_usecase

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"github.com/AleksandrMac/fileserver/internal/templates"
	"github.com/rs/zerolog/log"
)

func (x *EditorUsecase) EditHtml(w io.Writer, username, userId, filename string) error {

	ext := filepath.Ext(filename)
	docType := getDocType(ext)

	// Ключ документа — используем кастомный префикс, чтобы знать имя файла
	docKey := base64.URLEncoding.EncodeToString([]byte(filename))

	downloadURL := fmt.Sprintf("%s/%s", x.baseUrl, strings.TrimPrefix(filename, "/"))
	callbackURL := fmt.Sprintf("%s/track", x.baseUrl)

	data := map[string]any{
		"EditorToken": x.GenerateEditorToken(downloadURL, callbackURL, string(docKey)),
		"DocServer":   x.docServerUrl,
		"FileName":    filename,
		"FileType":    strings.TrimPrefix(ext, "."),
		"DocType":     docType,
		"DocKey":      docKey,
		"DownloadUrl": downloadURL,
		"CallbackUrl": callbackURL,
		"UserName":    username,
		"UserId":      userId,
	}

	err := template.Must(
		template.New("editor.html").
			ParseFS(templates.HTML, "html/editor.html")).
		Execute(w, data)
	if err != nil {
		log.Error().Err(err).Msg("failed parse editor.html")
	}
	return nil
}
