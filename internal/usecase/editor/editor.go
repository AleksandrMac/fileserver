package editor_usecase

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type EditorUsecase struct {
	jwtSecret    string
	docServerUrl string
	baseUrl      string
}

func NewEditorUsecase(jwtSecret, docServerUrl, baseUrl string) *EditorUsecase {
	return &EditorUsecase{
		jwtSecret:    jwtSecret,
		docServerUrl: docServerUrl,
		baseUrl:      baseUrl,
	}
}

func (x *EditorUsecase) GenerateEditorToken(docURL, callbackURL, docKey string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"document": map[string]any{
			"url": docURL,
			"key": docKey,
			"permissions": map[string]any{
				"modifyFilter": false,
			},
		},
		"editorConfig": map[string]any{
			"lang":        "ru",
			"mode":        "edit",
			"callbackUrl": callbackURL,
		},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(x.jwtSecret))
	if err != nil {
		return ""
	}
	return signed
}

func (x *EditorUsecase) VerifyEditorToken(tokenStr string) bool {
	// Валидация токена (без проверки payload — он может быть пустым)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(x.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return false
	}
	return true
}
