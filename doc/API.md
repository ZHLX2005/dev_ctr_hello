# File Server API Documentation

## Overview

A private file server for uploading and managing temporary files, optimized for LLM processing. Files have configurable expiration time, default is 1 hour.

**Server URL**: `http://localhost:8080`

**Version**: v1.0.0

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `AUTH_TOKEN` | Authentication token | `your-secret-token-change-me` |
| `DEFAULT_TTL` | Default file expiration time | `1h` |
| `STORAGE_DIR` | File storage directory | `/app/storage` |

---

## Authentication

### Bearer Token

Upload and delete endpoints require Bearer token authentication.

#### Request Header

| Header | Description |
|--------|-------------|
| `Authorization` | Bearer token (format: `Bearer <token>`) |

---

## API Endpoints

### 1. Health Check

**Request**
```
GET /health
```

**Response**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

### 2. Upload File

**Authentication Required**

**Request**
```
POST /api/v1/upload
Content-Type: multipart/form-data
Authorization: Bearer <your-token>
```

**Form Parameters**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file` | File | Yes | File to upload |
| `ttl` | String | No | Expiration time (e.g., `2h`, `30m`), uses default if not specified |

**Success Response** (200)
```json
{
  "id": "a1b2c3d4e5f6...",
  "name": "document.pdf",
  "size": 12345,
  "content_type": "application/pdf",
  "upload_time": "2025-12-25T10:30:00Z",
  "expires_at": "2025-12-25T11:30:00Z",
  "download_url": "http://localhost:8080/api/v1/download/a1b2c3d4e5f6..."
}
```

**Error Responses**
| Status | Description |
|--------|-------------|
| 400 | File parameter missing |
| 401 | Invalid or missing token |
| 500 | Internal server error |

---

### 3. Download File

**No Authentication Required**

**Request**
```
GET /api/v1/download/:id
```

**Path Parameters**
| Parameter | Description |
|-----------|-------------|
| `id` | File ID |

**Success Response** (200)
- Returns file binary content

**Response Headers**
| Header | Description |
|--------|-------------|
| `Content-Type` | File MIME type |
| `Content-Disposition` | Download disposition |
| `Content-Length` | File size |
| `X-File-Name` | Original filename |
| `X-Upload-Time` | Upload time |
| `X-Expires-At` | Expiration time |

**Error Responses**
| Status | Description |
|--------|-------------|
| 404 | File not found |
| 410 | File has expired |
| 500 | Internal server error |

---

### 4. Get File Metadata

**No Authentication Required**

**Request**
```
GET /api/v1/file/:id/metadata
```

**Path Parameters**
| Parameter | Description |
|-----------|-------------|
| `id` | File ID |

**Success Response** (200)
```json
{
  "id": "a1b2c3d4e5f6...",
  "name": "document.pdf",
  "size": 12345,
  "content_type": "application/pdf",
  "upload_time": "2025-12-25T10:30:00Z",
  "expires_at": "2025-12-25T11:30:00Z"
}
```

**Error Responses**
| Status | Description |
|--------|-------------|
| 404 | File not found |
| 410 | File has expired |
| 500 | Internal server error |

---

### 5. Delete File

**Authentication Required**

**Request**
```
DELETE /api/v1/file/:id
Authorization: Bearer <your-token>
```

**Path Parameters**
| Parameter | Description |
|-----------|-------------|
| `id` | File ID |

**Success Response** (200)
```json
{
  "message": "file deleted successfully"
}
```

**Error Responses**
| Status | Description |
|--------|-------------|
| 401 | Invalid or missing token |
| 500 | Internal server error |

---

## Usage Examples

### cURL Examples

#### Upload File
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer your-secret-token-change-me" \
  -F "file=@document.pdf" \
  -F "ttl=2h"
```

#### Download File
```bash
curl -O http://localhost:8080/api/v1/download/a1b2c3d4e5f6...
```

#### Get File Metadata
```bash
curl http://localhost:8080/api/v1/file/a1b2c3d4e5f6.../metadata
```

#### Delete File
```bash
curl -X DELETE http://localhost:8080/api/v1/file/a1b2c3d4e5f6... \
  -H "Authorization: Bearer your-secret-token-change-me"
```

### Go Client Example

```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

const (
    serverURL = "http://localhost:8080"
    authToken = "your-secret-token-change-me"
)

func main() {
    uploadFile("document.pdf")
}

func uploadFile(filePath string) {
    // Open file
    file, _ := os.Open(filePath)
    defer file.Close()

    // Create request body
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, _ := writer.CreateFormFile("file", filePath)
    io.Copy(part, file)
    writer.WriteField("ttl", "2h")
    writer.Close()

    // Create request
    req, _ := http.NewRequest("POST", serverURL+"/api/v1/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer "+authToken)

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
        return
    }
    defer resp.Body.Close()

    // Handle response...
}
```

### Python Client Example

```python
import requests

SERVER_URL = "http://localhost:8080"
AUTH_TOKEN = "your-secret-token-change-me"

# Upload file
with open('document.pdf', 'rb') as f:
    files = {'file': f}
    data = {'ttl': '2h'}
    headers = {'Authorization': f'Bearer {AUTH_TOKEN}'}

    response = requests.post(
        f'{SERVER_URL}/api/v1/upload',
        files=files,
        data=data,
        headers=headers
    )

    print(response.json())

# Download file
file_id = "a1b2c3d4e5f6..."
response = requests.get(f'{SERVER_URL}/api/v1/download/{file_id}')
with open('downloaded.pdf', 'wb') as f:
    f.write(response.content)
```

---

## Quick Start

### 1. Configure Token

Edit `.env` file and set your secure token:
```bash
AUTH_TOKEN=your-secure-random-token-here
```

### 2. Start Server

```bash
# Using Make
make server-start

# Or using Docker Compose
docker-compose up -d
```

### 3. Test API

```bash
# Health check
curl http://localhost:8080/health

# Upload file
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer your-secure-random-token-here" \
  -F "file=@test.txt"
```

---

## Notes

1. **Token Security**: Change the default `AUTH_TOKEN` to a secure random string before deployment
2. **File Expiration**: Files expire after 1 hour by default (configurable via `DEFAULT_TTL`)
3. **Storage Mount**: Mount the storage directory to host to persist data across container restarts
4. **Public Endpoints**: Download and metadata endpoints are public (no authentication required)

---

## Error Codes

| Status | Description |
|--------|-------------|
| 200 | Success |
| 204 | OPTIONS preflight success |
| 400 | Bad request |
| 401 | Unauthorized (invalid or missing token) |
| 404 | File not found |
| 410 | File has expired |
| 500 | Internal server error |
