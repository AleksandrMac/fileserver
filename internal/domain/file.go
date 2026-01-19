package domain

import "time"

type FileInfo struct {
	Name           string    `json:"name"`
	ModTime        time.Time `json:"mod_time"`
	SHA256Checksum string    `json:"sha256"`
}
