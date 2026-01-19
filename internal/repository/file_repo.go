package repository

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AleksandrMac/fileserver/internal/domain"
)

type FileRepository struct {
	storagePath string
}

func NewFileRepository(storagePath string) *FileRepository {
	err := os.MkdirAll(storagePath, 0755)
	if err != nil {
		panic("failed create FileRepository: " + err.Error())
	}
	return &FileRepository{storagePath: storagePath}
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

		files = append(files, domain.FileInfo{
			Name:    f.Name,
			ModTime: f.Modified,
		})
	}

	return files, nil
}

func (x *FileRepository) GetStorageSize() (int64, error) {
	var total int64

	err := filepath.Walk(x.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			total += info.Size()
		}

		return nil
	})

	return total, err
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
