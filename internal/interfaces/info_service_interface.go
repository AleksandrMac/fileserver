package interfaces

import "github.com/AleksandrMac/fileserver/internal/domain"

type InfoServiceInterface interface {
	GetInfo() *domain.ServiceInfo
}
