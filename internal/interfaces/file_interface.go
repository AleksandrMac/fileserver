package interfaces

import (
	"io"
	"os"

	"github.com/AleksandrMac/fileserver/internal/domain"
)

type FileRepo interface {
	GetFilePath(relPath string) (string, error)
	FileExists(path string) (bool, int64, error)
	SaveFile(path string, data io.Reader) error
	ListZipContents(zipPath string) ([]domain.FileInfo, error)
	GetStorageSize() (int64, error)
	ReadFile(path string) (*os.File, error)
	GetFileSize(path string) (int64, error)
}

type FileUsecase interface {
	GetFilePath(relPath string) (string, error)
	FileExists(path string) (bool, int64, error)
	SaveFile(path string, data io.Reader) error
	ListZipContents(zipPath string) ([]domain.FileInfo, error)
	GetStorageSize() (int64, error)
	ReadFile(path string) (io.ReadCloser, error)
	GetFileSize(path string) (int64, error)
}
