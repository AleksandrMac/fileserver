package repository

import (
	"testing"
)

func TestValidateAndResolvePath(t *testing.T) {
	root := "/safe/storage"

	tests := []struct {
		name     string
		userPath string
		wantErr  bool
		wantPath string // ожидаемый полный путь (если нет ошибки)
	}{
		// Успешные случаи
		{
			name:     "normal file",
			userPath: "folder/file.txt",
			wantErr:  false,
			wantPath: "\\safe\\storage\\folder\\file.txt",
		},
		{
			name:     "nested path",
			userPath: "a/b/c/d.txt",
			wantErr:  false,
			wantPath: "\\safe\\storage\\a\\b\\c\\d.txt",
		},
		{
			name:     "path with slashes",
			userPath: "/docs/report.pdf",
			wantErr:  false,
			wantPath: "\\safe\\storage\\docs\\report.pdf",
		},
		{
			name:     "path with dot in name",
			userPath: "file.tar.gz",
			wantErr:  false,
			wantPath: "\\safe\\storage\\file.tar.gz",
		},
		{
			name:     "root-level file",
			userPath: "readme.md",
			wantErr:  false,
			wantPath: "\\safe\\storage\\readme.md",
		},

		// Ошибки: path traversal
		{
			name:     "basic path traversal",
			userPath: "../etc/passwd",
			wantErr:  false,
			wantPath: "\\safe\\storage\\etc\\passwd",
		},
		{
			name:     "absolute path traversal",
			userPath: "/../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "encoded traversal",
			userPath: "..%2F..%2Fsecret",
			wantErr:  false, // НЕ декодируем URL здесь — это делается выше!
			// Примечание: URL-декодирование должно происходить до вызова этой функции.
			// Иначе "%2F" станет "/", и путь изменится.
			// В данном тесте мы проверяем ТОЛЬКО путь, уже прошедший URL-декодирование.
			wantPath: "\\safe\\storage\\..%2F..%2Fsecret",
		},
		{
			name:     "traversal with clean",
			userPath: "a/../../secret",
			wantErr:  false,
			wantPath: "\\safe\\storage\\secret",
		},

		// Пограничные случаи
		{
			name:     "double slash",
			userPath: "//docs//file.txt",
			wantErr:  false,
			wantPath: "\\safe\\storage\\docs\\file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &FileRepository{storagePath: root}
			got, err := repo.validateAndCleanPath(tt.userPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndCleanPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantPath {
				t.Errorf("validateAndCleanPath() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}
