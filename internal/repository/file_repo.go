package repository

import (
	"io"
	"os"

	"github.com/AleksandrMac/fileserver/internal/domain"
)

type FileRepository struct {
	storagePath string
}

func NewFileRepository(storagePath string) *FileRepository {
	return &FileRepository{storagePath: storagePath}
}

func (x *FileRepository) GetFilePath(relPath string) string {
	return ""
}

func (x *FileRepository) FileExists(path string) (bool, int64, error) {
	return false, 0, nil
}

func (x *FileRepository) SaveFile(path string, data io.Reader) error {
	return nil
}

func (x *FileRepository) ListZipContents(zipPath string) ([]domain.FileInfo, error) {
	return nil, nil
}

func (x *FileRepository) GetStorageSize() (int64, error) {
	return 0, nil
}

func (x *FileRepository) ReadFile(path string) (*os.File, error) {
	return nil, nil
}

func (x *FileRepository) GetFileSize(path string) (int64, error) {
	return 0, nil
}
