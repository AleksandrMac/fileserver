package usecase

import (
	"io"
	"os"

	"github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/AleksandrMac/fileserver/internal/interfaces"
)

type FileUsecase struct {
	fileRepo interfaces.FileRepo
}

func NewFileUseCase(fileRepo interfaces.FileRepo) *FileUsecase {
	return &FileUsecase{
		fileRepo: fileRepo,
	}
}

func (x *FileUsecase) GetFullPath(relPath string) (string, error) {
	return x.fileRepo.GetFullPath(relPath)
}

func (x *FileUsecase) FileInfo(path string) (os.FileInfo, error) {
	return x.fileRepo.FileInfo(path)
}

func (x *FileUsecase) SaveFile(path string, data io.Reader) error {
	return x.fileRepo.SaveFile(path, data)
}

func (x *FileUsecase) List(path string) ([]domain.FileInfo, error) {
	return x.fileRepo.List(path)
}

func (x *FileUsecase) ListZipContents(zipPath string) ([]domain.FileInfo, error) {
	return x.fileRepo.ListZipContents(zipPath)
}

func (x *FileUsecase) GetStorageInfo() (*domain.StorageInfo, error) {
	return x.fileRepo.GetStorageInfo()
}

func (x *FileUsecase) ReadFile(path string) (io.ReadCloser, error) {
	return x.fileRepo.ReadFile(path)
}

func (x *FileUsecase) GetFileSize(path string) (int64, error) {
	return x.fileRepo.GetFileSize(path)
}
