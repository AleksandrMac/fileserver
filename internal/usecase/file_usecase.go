package usecase

import (
	"io"

	"github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/AleksandrMac/fileserver/internal/interfaces"
)

type FileUsecase struct {
	fileRepo interfaces.FileRepo
}

func NewFileUseCase(fileRepo interfaces.FileRepo) *FileUsecase {
	return &FileUsecase{fileRepo: fileRepo}
}

func (x *FileUsecase) GetFilePath(relPath string) (string, error) {
	return x.fileRepo.GetFilePath(relPath)
}

func (x *FileUsecase) FileExists(path string) (bool, int64, error) {
	return x.fileRepo.FileExists(path)
}

func (x *FileUsecase) SaveFile(path string, data io.Reader) error {
	return x.fileRepo.SaveFile(path, data)
}

func (x *FileUsecase) ListZipContents(zipPath string) ([]domain.FileInfo, error) {
	return x.fileRepo.ListZipContents(zipPath)
}

func (x *FileUsecase) GetStorageSize() (int64, error) {
	return x.fileRepo.GetStorageSize()
}

func (x *FileUsecase) ReadFile(path string) (io.ReadCloser, error) {
	return x.fileRepo.ReadFile(path)
}

func (x *FileUsecase) GetFileSize(path string) (int64, error) {
	return x.fileRepo.GetFileSize(path)
}
