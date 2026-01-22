package domain

type ServiceInfo struct {
	Version   string       `json:"version"`
	Commit    string       `json:"commit"`
	BuildTime string       `json:"build_time"`
	Port      string       `json:"port"`
	Storage   *StorageInfo `json:"storage"`
}

type StorageInfo struct {
	Path       string `json:"path"`
	TotalFiles int64  `json:"total_files"`
	TotalSize  int64  `json:"total_size_bytes"`
}
