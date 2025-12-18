package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TotalStorageSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "fileserver_total_storage_bytes",
		Help: "Total size of all stored files in bytes",
	})

	RequesCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fileserver_requests_total",
		Help: "Total number of HTTP request",
	}, []string{"method", "path", "status"})

	BytesDownloaded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fileserver_bytes_downloaded_total",
		Help: "Total number of bytes downloaded",
	})

	BytesUploaded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "fileserver_bytes_uploaded_total",
		Help: "Total number of bytes uploaded",
	})
)
