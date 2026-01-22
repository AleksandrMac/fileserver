package usecase

import (
	"github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/AleksandrMac/fileserver/internal/interfaces"
)

type InfoService struct {
	version   string
	commit    string
	buildTime string
	port      string
	storage   interfaces.StorageInfo
}

func NewInfoService(version, commit, buildTime, port string, storage interfaces.StorageInfo) *InfoService {
	return &InfoService{
		version:   version,
		commit:    commit,
		buildTime: buildTime,
		port:      port,
		storage:   storage,
	}
}

func (s *InfoService) GetInfo() *domain.ServiceInfo {
	storage, _ := s.storage.GetStorageInfo()

	return &domain.ServiceInfo{
		Version:   s.version,
		Commit:    s.commit,
		BuildTime: s.buildTime,
		Storage:   storage,
		Port:      s.port,
	}
}
