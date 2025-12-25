# 文件服务器 API 文档

## 概述

这是一个私有文件服务器，用于上传和管理临时文件，支持大模型识别。文件具有可配置的过期时间，默认为 1 小时。

**服务地址**: `http://localhost:8080`

**版本**: v1.0.0

---

## 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | `8080` |
| `DEFAULT_TTL` | 文件默认过期时间 | `1h` |
| `PUBLIC_KEY_PATH` | 公钥文件路径 | `/app/keys/public.pem` |
| `STORAGE_DIR` | 文件存储目录 | `/app/storage` |

---

## 认证机制

### 签名验证

上传和删除接口需要使用 RSA 签名进行认证。

#### 签名生成方式

待签名字符串格式:
```
METHOD:PATH:TIMESTAMP
```

- `METHOD`: HTTP 方法（如 `POST`）
- `PATH`: 请求路径（如 `/api/v1/upload`）
- `TIMESTAMP`: RFC3339 格式的时间戳

#### 请求头

| 头名称 | 说明 |
|--------|------|
| `X-Signature` | Base64 编码的 RSA 签名 |
| `X-Timestamp` | RFC3339 格式的时间戳 |

#### 签名时间容差

签名时间容差为 ±5 分钟，超过该时间范围的请求将被拒绝。

---

## API 接口

### 1. 健康检查

**请求**
```
GET /health
```

**响应**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

### 2. 上传文件

**需要认证**

**请求**
```
POST /api/v1/upload
Content-Type: multipart/form-data
X-Signature: <签名>
X-Timestamp: <时间戳>
```

**表单参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | File | 是 | 要上传的文件 |
| `ttl` | String | 否 | 过期时间（如 `2h`、`30m`），默认使用配置的默认值 |

**成功响应** (200)
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

**错误响应**
| 状态码 | 说明 |
|--------|------|
| 400 | 文件参数缺失 |
| 401 | 签名验证失败或时间戳无效 |
| 500 | 服务器内部错误 |

---

### 3. 下载文件

**不需要认证**

**请求**
```
GET /api/v1/download/:id
```

**路径参数**
| 参数 | 说明 |
|------|------|
| `id` | 文件 ID |

**成功响应** (200)
- 返回文件二进制内容

**响应头**
| 头名称 | 说明 |
|--------|------|
| `Content-Type` | 文件 MIME 类型 |
| `Content-Disposition` | 文件下载 disposition |
| `Content-Length` | 文件大小 |
| `X-File-Name` | 原始文件名 |
| `X-Upload-Time` | 上传时间 |
| `X-Expires-At` | 过期时间 |

**错误响应**
| 状态码 | 说明 |
|--------|------|
| 404 | 文件不存在 |
| 410 | 文件已过期 |
| 500 | 服务器内部错误 |

---

### 4. 获取文件元数据

**不需要认证**

**请求**
```
GET /api/v1/file/:id/metadata
```

**路径参数**
| 参数 | 说明 |
|------|------|
| `id` | 文件 ID |

**成功响应** (200)
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

**错误响应**
| 状态码 | 说明 |
|--------|------|
| 404 | 文件不存在 |
| 410 | 文件已过期 |
| 500 | 服务器内部错误 |

---

### 5. 删除文件

**需要认证**

**请求**
```
DELETE /api/v1/file/:id
X-Signature: <签名>
X-Timestamp: <时间戳>
```

**路径参数**
| 参数 | 说明 |
|------|------|
| `id` | 文件 ID |

**成功响应** (200)
```json
{
  "message": "file deleted successfully"
}
```

**错误响应**
| 状态码 | 说明 |
|--------|------|
| 401 | 签名验证失败或时间戳无效 |
| 500 | 服务器内部错误 |

---

## 使用示例

### cURL 示例

#### 上传文件
```bash
# 首先计算签名（需要私钥）
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
SIGN_DATA="POST:/api/v1/upload:${TIMESTAMP}"
SIGNATURE=$(echo -n "$SIGN_DATA" | openssl dgst -sha256 -sign private.pem -base64)

# 上传文件
curl -X POST http://localhost:8080/api/v1/upload \
  -H "X-Signature: $SIGNATURE" \
  -H "X-Timestamp: $TIMESTAMP" \
  -F "file=@document.pdf" \
  -F "ttl=2h"
```

#### 下载文件
```bash
curl -O http://localhost:8080/api/v1/download/a1b2c3d4e5f6...
```

#### 获取文件元数据
```bash
curl http://localhost:8080/api/v1/file/a1b2c3d4e5f6.../metadata
```

### Go 客户端示例

```go
package main

import (
    "bytes"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "time"

    "dev_ctr_hello/pkg/auth"
)

func main() {
    // 加载私钥
    privateKey, _ := auth.LoadPrivateKey("keys/private.pem")

    // 上传文件
    uploadFile(privateKey)
}

func uploadFile(privateKey *rsa.PrivateKey) {
    // 准备文件
    file, _ := os.Open("document.pdf")
    defer file.Close()

    // 创建请求体
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, _ := writer.CreateFormFile("file", "document.pdf")
    io.Copy(part, file)
    writer.WriteField("ttl", "2h")
    writer.Close()

    // 生成签名
    timestamp := time.Now().Format(time.RFC3339)
    signData := auth.GenerateSignData("POST", "/api/v1/upload", "", timestamp, "")
    signature, _ := auth.SignWithKey([]byte(signData), privateKey)

    // 发送请求
    req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("X-Signature", signature)
    req.Header.Set("X-Timestamp", timestamp)

    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()

    // 处理响应...
}
```

### Python 客户端示例

```python
import requests
import time
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.backends import default_backend
import base64

# 加载私钥
with open('keys/private.pem', 'rb') as f:
    private_key = serialization.load_pem_private_key(
        f.read(),
        password=None,
        backend=default_backend()
    )

# 生成签名
timestamp = time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())
sign_data = f'POST:/api/v1/upload:{timestamp}'
signature = private_key.sign(
    sign_data.encode(),
    padding.PKCS1v15(),
    hashes.SHA256()
)
signature_b64 = base64.b64encode(signature).decode()

# 上传文件
files = {'file': open('document.pdf', 'rb')}
data = {'ttl': '2h'}
headers = {
    'X-Signature': signature_b64,
    'X-Timestamp': timestamp
}

response = requests.post(
    'http://localhost:8080/api/v1/upload',
    files=files,
    data=data,
    headers=headers
)

print(response.json())
```

---

## 快速开始

### 1. 生成密钥

```bash
# Windows PowerShell
powershell -ExecutionPolicy Bypass -File scripts\generate_keys.ps1

# Linux/Mac/Git Bash
bash scripts/generate_keys.sh
```

### 2. 启动服务

```bash
# 使用 Make
make server-start

# 或使用 Docker Compose
docker-compose up -d
```

### 3. 测试接口

```bash
# 健康检查
curl http://localhost:8080/health
```

---

## 注意事项

1. **私钥安全**: 私钥文件 (`private.pem`) 仅在客户端使用，切勿上传到服务器或提交到版本控制
2. **公钥部署**: 公钥文件 (`public.pem`) 需要放置在服务器的 `/app/keys/` 目录
3. **文件过期**: 文件默认 1 小时后过期，过期后无法访问
4. **存储挂载**: 文件存储目录应挂载到宿主机，避免容器重启丢失数据
5. **时间同步**: 客户端和服务器时间应保持同步，避免签名验证失败

---

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 204 | OPTIONS 预检请求成功 |
| 400 | 请求参数错误 |
| 401 | 认证失败（签名无效或时间戳过期） |
| 404 | 文件不存在 |
| 410 | 文件已过期 |
| 500 | 服务器内部错误 |
