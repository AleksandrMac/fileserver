# 📁 File Server — Lightweight API for File Management with Archive Inspection

A minimal, secure, and observable file server written in Go, designed for Kubernetes environments. Supports file upload/download, on-the-fly ZIP archive metadata inspection, Prometheus metrics, health checks, and graceful shutdown — all configurable via environment variables.

Built with **Clean Architecture**, **go-chi**, and production-grade practices.

---

## ✨ Features

- ✅ **File upload** with `X-API-Key` authorization  
- ✅ **File download** via `GET /<path>`  
- ✅ **ZIP archive inspection**: `GET /archive.zip?meta=true` returns JSON list of files, modification times, and SHA256 hashes  
- ✅ **HTTP methods**: `GET`, `HEAD`, `OPTIONS` for archives; `POST` for uploads  
- ✅ **Prometheus metrics**:
  - Total storage size (`fileserver_total_storage_bytes`)
  - Request count by method/path/status (`fileserver_requests_total`)
  - Bytes downloaded/uploaded (`fileserver_bytes_downloaded_total`, `fileserver_bytes_uploaded_total`)
- ✅ **Kubernetes-ready**:
  - `/health` (liveness probe)
  - `/ready` (readiness probe)
  - Graceful shutdown on `SIGTERM`
- ✅ **Secure**:
  - Path traversal protection
  - No cloud dependencies — works with local filesystem
- ✅ **Configurable** via environment variables
- ✅ **Lightweight**: single Go binary (~15 MB)

---

## 🚀 Quick Start

### 1. Build & Run

```bash
# Build
go build -o fileserver cmd/fileserver/main.go

# Set environment variables
export API_KEY=your-secret-key
export STORAGE_PATH=./data
export PORT=8080

# Run
./fileserver
```

### 2. Try It
```bash
# Upload a file
curl -H "X-API-Key: your-secret-key" \
     -F "file=@document.pdf" \
     "http://localhost:8080/upload?path=docs/document.pdf"

# Download a file
curl -O http://localhost:8080/docs/document.pdf

# Inspect ZIP archive
curl "http://localhost:8080/data.zip?meta=true"

# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics
```

---

## ⚙️ Configuration (Environment Variables)

| Month    | Savings |
| -------- | ------- |
| January  | $250    |
| February | $80     |
| March    | $420    |


| Variable | Required | Default | Description|
| -------- | -------- | ------- | ---------- |
| API_KEY  | ✅ Yes  |—         | API key for upload authorization (X-API-Key header)|
| STORAGE_PATH | ❌ No | ./storage | Root directory for stored files "|
| PORT | ❌ No | 8080 | HTTP server port |

> 🔐 `Security Note`: Never expose this service publicly without a reverse proxy (e.g., NGINX, Traefik) handling TLS and network policies.

---

## 📊 Metrics (Prometheus)

Expose metrics at http://<host>:<port>/metrics. Example:

```prometheus
fileserver_total_storage_bytes 204800
fileserver_requests_total{method="GET",path="/data.zip",status="200"} 5
fileserver_bytes_downloaded_total 1024000
fileserver_bytes_uploaded_total 512000
```

Useful for alerting on storage growth or traffic spikes.

---

## 🧪 API Reference

`POST /upload?path=<rel_path>`

Upload file to store

- Headers: `X-API-Key: <your_key>`
- Body: `multipart/form-data` with `file` field
- Response: `201 Created` on success
  
`GET /<file_path>`

Downloads file

- Supports HEAD and OPTIONS

`GET /<archive.zip>?meta=true`

Get metadata in archive

Returns JSON array:

```json
[
  {
    "name": "file.txt",
    "mod_time": "2024-12-01T10:00:00Z",
    "sha256": "a1b2c3..."
  }
]
```

`GET /health`

Liveness probe → returns 200 OK

`GET /ready`

Readiness probe → returns 200 OK

`GET /metrics`

Prometheus metrics endpoint

## 🧱 Architecture

Follows Clean Architecture principles:

```
main
 └── delivery/http (go-chi handlers, middleware)
 └── usecase (business logic)
 └── repository (file system abstraction)
 └── domain (entities, no dependencies)
```
Easy to extend (e.g., switch to S3 by implementing new repository).

---

### 📜 License

MIT