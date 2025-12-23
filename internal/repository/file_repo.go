package repository

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/AleksandrMac/fileserver/internal/domain"
)

type FileRepository struct {
	storagePath string
}

func NewFileRepository(storagePath string) *FileRepository {
	return &FileRepository{storagePath: storagePath}
}

func (x *FileRepository) GetFilePath(relPath string) string {
	return filepath.Join(x.storagePath, relPath)
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

func (x *FileRepository) SaveFile(path string, data io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, data)

	return err
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
