package repository

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/charmap"

	"github.com/AleksandrMac/fileserver/internal/domain"
)

type FileRepository struct {
	storagePath      string
	fallbackEncoding *charmap.Charmap
}

func NewFileRepository(storagePath string) *FileRepository {
	err := os.MkdirAll(storagePath, 0755)
	if err != nil {
		panic("failed create FileRepository: " + err.Error())
	}
	path, err := filepath.Abs(storagePath)
	if err != nil {
		panic("failed get absolute path: " + err.Error())
	}
	return &FileRepository{
		storagePath:      path,
		fallbackEncoding: charmap.CodePage866,
	}
}

func (x *FileRepository) GetFilePath(relPath string) (string, error) {
	return x.validateAndCleanPath(relPath)
}

func (x *FileRepository) FileExists(path string) (bool, int64, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return !info.IsDir(), info.Size(), nil
}

func (x *FileRepository) SaveFile(fullPath string, data io.Reader) error {
	if !strings.HasPrefix(fullPath, x.storagePath) {
		return errors.New("failed path, want absoulute path.")
	}

	// 1. создаем все директории в пути
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// 2. создаем временный файл для записи
	tempFile, err := os.CreateTemp(filepath.Dir(fullPath), ".tmp_")
	if err != nil {
		return err
	}

	// 3. пишем данные во временный файл
	_, err = io.Copy(tempFile, data)
	closeErr := tempFile.Close()
	if err != nil || closeErr != nil {
		os.Remove(tempFile.Name())
		if closeErr != nil {
			return closeErr
		}
		return err
	}

	// 4. Атомарно переименовываем (в Linux/Mac — это atomic)
	return os.Rename(tempFile.Name(), fullPath)
}

func (x *FileRepository) ListZipContents(zipPath string) ([]domain.FileInfo, error) {
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	var files []domain.FileInfo
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// --- Декодирование имени файла ---
		filename := f.Name
		if f.Flags&0x800 == 0 {
			// Флаг UTF-8 НЕ установлен → предполагаем локальную кодировку
			if decoded, err := decodeString(filename, x.fallbackEncoding); err == nil {
				filename = decoded
			} else {
				// Если декодирование сломалось — оставляем как есть (лучше битое имя, чем падение)
				// Можно логировать: log.Warn().Str("original", f.Name).Msg("failed to decode filename")
			}
		}

		files = append(files, domain.FileInfo{
			Name:    filename,
			ModTime: f.Modified,
		})
	}

	return files, nil
}

func (x *FileRepository) GetStorageInfo() (*domain.StorageInfo, error) {
	totalFiles := int64(0)
	totalSize := int64(0)

	if err := filepath.Walk(x.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}

		return nil
	}); err != nil {
		return &domain.StorageInfo{Path: x.storagePath}, err
	}

	return &domain.StorageInfo{
		Path:       x.storagePath,
		TotalFiles: totalFiles,
		TotalSize:  totalSize,
	}, nil
}

func (x *FileRepository) ReadFile(path string) (*os.File, error) {
	return os.Open(path)
}

func (x *FileRepository) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func (x *FileRepository) validateAndCleanPath(path string) (string, error) {
	// Нормализуем путь
	cleanPath := filepath.Clean("/" + path)

	if cleanPath == "\\" {
		return "", errors.New("invalid path: path is empty")
	}

	// Запрещаем подъемы выше корня
	if strings.Contains(cleanPath, "..\\") {
		return "", errors.New("invalid path: contains '..'")
	}

	// возвращаем полный путь
	return filepath.Join(x.storagePath, cleanPath), nil
}

// decodeString преобразует строку из заданной кодировки в UTF-8
func decodeString(s string, enc *charmap.Charmap) (string, error) {
	decoder := enc.NewDecoder()
	decoded, err := decoder.String(s)
	if err != nil {
		return "", err
	}
	return decoded, nil
}
