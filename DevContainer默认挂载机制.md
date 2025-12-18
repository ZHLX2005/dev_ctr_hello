# DevContainer 默认挂载机制详解

## 概述

本文档详细说明了 VSCode DevContainer 的默认文件挂载机制，以及它如何将本地源代码文件挂载到容器中。

## 默认挂载机制

### 1. 自动挂载行为

当 DevContainer 配置文件（`.devcontainer/devcontainer.json`）中**没有明确指定挂载配置**时，VSCode 会使用默认的挂载规则：

#### 默认配置
```json
// VSCode DevContainer 内部的默认配置（用户不可见）
{
  "workspaceFolder": "/workspace",
  "workspaceMount": "src=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached"
}
```

### 2. 挂载参数说明

#### workspaceMount 参数解析
- `src=${localWorkspaceFolder}`：源路径（本地项目目录）
  - Windows: `d:\code\a_go\proj\dev_ctr_hello`
  - Linux/Mac: `/home/user/projects/dev_ctr_hello`

- `target=/workspace`：目标路径（容器内挂载点）

- `type=bind`：挂载类型
  - `bind`：直接绑定挂载，实时同步

- `consistency=cached`：一致性模式
  - `cached`：优化性能，文件变更可能有轻微延迟（默认）
  - `consistent`：实时同步，性能较差
  - `delegated`：优先容器端性能

### 3. 实际执行的 Docker 命令

#### Windows 环境
```bash
docker run -it --name dev_ctr_hello-12345 \
  -v "d:\\code\\a_go\\proj\\dev_ctr_hello:/workspace:cached" \
  -p 8080:8080 \
  -e GOPROXY=https://goproxy.cn,direct \
  --user vscode \
  golang:1.25-bookworm \
  sleep infinity
```

#### Linux/Mac 环境
```bash
docker run -it --name dev_ctr_hello-12345 \
  -v "/home/user/projects/dev_ctr_hello:/workspace:cached" \
  -p 8080:8080 \
  -e GOPROXY=https://goproxy.cn,direct \
  --user vscode \
  golang:1.25-bookworm \
  sleep infinity
```

## 文件系统映射

### 本地到容器的文件映射

```
本地文件系统                             容器内文件系统
├── d:\code\a_go\proj\dev_ctr_hello\      └── /workspace/
    ├── .devcontainer/  ❌                 → .devcontainer/ (忽略)
    ├── .vscode/       ⚠️                 → .vscode/ (可选挂载)
    ├── .gitignore                       → .gitignore
    ├── .air.toml                        → .air.toml
    ├── README.md                        → README.md
    ├── go.mod                           → go.mod
    ├── go.sum                           → go.sum
    ├── main.go                          → main.go
    ├── Makefile                         → Makefile
    ├── scripts/                         → scripts/
    │   ├── run.sh                       → scripts/run.sh
    │   └── run.bat                      → scripts/run.bat
    └── bin/                             → bin/
```

### 文件同步规则

#### ✅ 会自动挂载的文件
- 所有项目源代码文件
- 配置文件（Makefile, .air.toml 等）
- 用户创建的目录和文件
- 构建输出目录（bin/, tmp/ 等）

#### ❌ 不会挂载的文件/目录
- `.devcontainer/` 目录（DevContainer 自动忽略）
- 容器系统文件（/usr, /bin, /etc 等）
- Go 模块缓存（通常在 `/go/pkg/mod`）
- VSCode Server 文件（在 `/home/vscode/.vscode-server`）

## 性能优化

### 1. 一致性模式选择

#### Windows WSL2 推荐
```json
{
  "mounts": [
    "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached"
  ]
}
```

#### 需要实时同步的场景
```json
{
  "mounts": [
    "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=consistent"
  ]
}
```

### 2. 排除不必要的文件

创建 `.dockerignore` 文件：
```dockerignore
# 排除大文件和缓存
node_modules/
.git/
*.log
.DS_Store
*.tmp

# 排除临时文件
tmp/
.cache/
.vscode/cache/
```

## 自定义挂载配置

### 1. 显式配置挂载

```json
{
  "workspaceFolder": "/workspace",
  "workspaceMount": "src=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached"
}
```

### 2. 添加额外的挂载点

```json
{
  "mounts": [
    // 项目文件（默认）
    "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached",

    // Go 缓存（提高构建速度）
    "source=dev-ctr-hello-go-cache,target=/go/pkg/mod,type=volume",

    // 本地配置文件
    "source=${localWorkspaceFolder}/.env,target=/workspace/.env,type=bind,consistency=cached",

    // 数据持久化
    "source=dev-ctr-hello-data,target=/app/data,type=volume"
  ]
}
```

### 3. 挂载本地 .vscode 配置

```json
{
  "mounts": [
    "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached",
    "source=${localWorkspaceFolder}/.vscode,target=/workspace/.vscode,type=bind"
  ]
}
```

## 验证挂载

### 1. 查看挂载信息
在容器内执行：
```bash
# 查看所有挂载点
mount | grep workspace

# 查看文件系统
df -h

# 查看目录内容
ls -la /workspace

# 检查挂载类型
mountpoint /workspace
```

### 2. 测试文件同步
```bash
# 在容器内创建文件
echo "test" > /workspace/test.txt

# 检查本地是否出现文件
ls -la test.txt

# 反向测试
# 在本地创建文件
# 在容器内查看文件是否出现
```

## 常见问题

### 1. 文件不同步

**问题**：修改本地文件，容器内看不到变化

**解决方案**：
```bash
# 检查挂载状态
docker inspect container_name | grep Mounts

# 重启 Docker Desktop（Windows）
# 或重启 Docker 服务（Linux）
```

### 2. 权限问题

**问题**：容器内无法写入文件

**解决方案**：
```json
{
  "remoteUser": "vscode",
  "runArgs": ["--userns=keep-id"]
}
```

或在 Dockerfile 中：
```dockerfile
RUN chown -R vscode:vscode /workspace
USER vscode
```

### 3. 性能问题

**问题**：文件操作很慢

**解决方案**：
```json
{
  "mounts": [
    "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached"
  ],
  "runArgs": [
    // Windows 特定优化
    "--mount", "type=volume,source=vscode-git,target=/tmp/vscode-git"
  ]
}
```

## 最佳实践

### 1. 保持默认配置
对于大多数项目，默认挂载机制已经足够好：
```json
// 最小配置
{
  "name": "Go DevContainer",
  "build": {
    "dockerfile": "Dockerfile",
    "context": ".."
  }
}
```

### 2. 按需自定义
只在必要时才自定义挂载：
- 需要持久化数据时使用 volume
- 需要访问特定本地路径时使用 bind mount
- 性能优化时才调整 consistency 模式

### 3. 考虑跨平台兼容性
```json
{
  "mounts": [
    "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached"
  ],
  "runArgs": [
    // Windows 特定
    "${localEnv:VSCODE_WSL_BIND_MOUNT_ARGS}"
  ]
}
```

## 总结

DevContainer 的默认挂载机制提供了一个简单而强大的文件同步方案：

1. **自动挂载**：无需配置即可使用
2. **实时同步**：修改立即反映（取决于 consistency 模式）
3. **跨平台支持**：Windows/Linux/Mac 一致体验
4. **性能优化**：cached 模式平衡性能和实时性

默认配置适合 90% 的使用场景，只在特殊需求时才需要自定义配置。