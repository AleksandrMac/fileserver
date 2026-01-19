package interfaces

import (
	"io"
	"os"

	"github.com/AleksandrMac/fileserver/internal/domain"
)

type StorageInfo interface {
	// GetStorageInfo возвращает информацию о хранилище
	GetStorageInfo() (*domain.StorageInfo, error)
}

type FileRepo interface {
	StorageInfo
	GetFilePath(relPath string) (string, error)
	FileExists(path string) (bool, int64, error)
	SaveFile(path string, data io.Reader) error
	ListZipContents(zipPath string) ([]domain.FileInfo, error)
	ReadFile(path string) (*os.File, error)
	GetFileSize(path string) (int64, error)
}

type FileUsecase interface {
	StorageInfo
	GetFilePath(relPath string) (string, error)
	FileExists(path string) (bool, int64, error)
	SaveFile(path string, data io.Reader) error
	ListZipContents(zipPath string) ([]domain.FileInfo, error)
	ReadFile(path string) (io.ReadCloser, error)
	GetFileSize(path string) (int64, error)
}
