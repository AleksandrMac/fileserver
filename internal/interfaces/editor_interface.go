package interfaces

import "io"

type EditorUsecase interface {
	GenerateEditorToken(docURL, callbackURL, docKey string) string
	VerifyEditorToken(token string) bool
	EditHtml(w io.Writer, userName, userId, fileName string) error
}
