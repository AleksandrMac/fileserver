package domain

import "time"

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}
