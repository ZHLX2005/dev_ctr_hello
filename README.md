# Gin DevContainer 示例项目

这是一个使用 VSCode DevContainer 搭建的 Go Gin 框架示例项目。

## 快速开始

### 1. 使用 VSCode 打开项目
```bash
code .
```

### 2. 在 DevContainer 中打开
1. 打开 VSCode 后，右下角会提示 "Reopen in Container"
2. 点击该按钮，等待容器构建完成
3. 或者按 `F1`，输入 `Dev Containers: Reopen in Container`

### 3. 运行项目

#### 使用 Make（推荐）
```bash
# 查看所有可用命令
make help

# 运行项目
make run

# 开发模式（热重载）
make run-dev

# 构建项目
make build

# 运行测试
make test

# 生成测试覆盖率报告
make test-cover

# 清理构建文件
make clean
```

#### 直接使用 Go 命令
```bash
# 直接运行
go run main.go

# 或使用 air 实现热重载
air
```

### 4. 访问应用
- 应用地址：http://localhost:8080
- API 文档：http://localhost:8080/api/v1

## 项目结构

```
dev_ctr_hello/
├── .devcontainer/            # DevContainer 配置
│   ├── devcontainer.json     # VSCode DevContainer 配置
│   └── Dockerfile           # 自定义 Docker 镜像
├── .vscode/                  # VSCode 配置
│   ├── launch.json          # 调试配置
│   └── tasks.json           # 任务配置
├── scripts/                  # 启动脚本
│   ├── run.sh               # Linux/Mac 启动脚本
│   └── run.bat              # Windows 启动脚本
├── main.go                   # 主入口文件
├── Makefile                  # Make 构建配置
├── go.mod                    # Go 模块文件
├── go.sum                    # 依赖锁定文件
├── .air.toml                 # Air 热重载配置
├── .gitignore                # Git 忽略文件
└── README.md                 # 项目说明文档
```

## API 端点

### 基础端点
- `GET /` - 欢迎页面
- `GET /ping` - Ping 测试
- `GET /health` - 健康检查

### API v1 端点
- `GET /api/v1/hello/:name` - 问候 API
- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取特定用户
- `POST /api/v1/users` - 创建新用户

## Make 命令详解

项目使用 Make 来管理常用的开发任务。运行 `make help` 查看所有可用命令。

### 核心命令
```bash
# 初始化项目（下载依赖、安装开发工具）
make init

# 开发相关
make run          # 运行应用
make run-dev      # 开发模式（热重载）
make build        # 构建应用到 bin/main
make build-linux  # 构建 Linux 二进制文件

# 测试相关
make test         # 运行测试
make test-cover   # 生成测试覆盖率报告
make test-bench   # 运行基准测试

# 代码质量
make fmt          # 格式化代码
make vet          # 代码静态检查
make lint         # 代码检查（需要 golangci-lint）
make check        # 运行所有代码检查

# 依赖管理
make deps         # 下载和管理依赖

# Docker
make docker-build # 构建 Docker 镜像
make docker-run   # 运行 Docker 容器

# 发布
make release      # 准备发布版本
```

### 高级用法
```bash
# 自定义版本号构建
VERSION=v1.0.1 make build

# 清理所有生成文件
make clean

# 安装到 GOPATH/bin
make install
```

## 开发工具

项目已配置以下开发工具：
- **gopls** - Go 语言服务器
- **air** - 热重载工具
- **golangci-lint** - 代码检查工具
- **swag** - API 文档生成工具

### 使用 air 进行热重载
```bash
# 安装 air
make dev-setup  # 包含安装 air

# 或单独安装
go install github.com/cosmtrek/air@latest

# 运行 air
air
# 或
make run-dev
```

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行基准测试
go test -bench=. ./...
```

## 调试配置

VSCode 已配置调试功能：
1. 在 `main.go` 中设置断点
2. 按 `F5` 启动调试
3. 或者点击 Run and Debug 面板中的 "Launch Package"

## 自定义配置

### 添加新的扩展
编辑 `.devcontainer/devcontainer.json`：
```json
"extensions": [
    "golang.go",
    "your-extension-name"
]
```

### 安装额外的 Go 工具
在 `.devcontainer/Dockerfile` 中添加：
```dockerfile
RUN go install github.com/example/tool@latest
```

## 常见问题

### Q: 如何更新 Go 版本？
A: 修改 `.devcontainer/devcontainer.json` 中的 `image` 字段。

### Q: 如何添加新的端口？
A: 在 `devcontainer.json` 中添加：
```json
"forwardPorts": [8080, 3000, 新端口]
```

### Q: 如何修改启动命令？
A: 修改 `main.go` 中的 `r.Run()` 端口号。

## 扩展项目

### 添加数据库
1. 修改 `.devcontainer/docker-compose.yml` 添加数据库服务
2. 在 `devcontainer.json` 中配置服务依赖

### 添加前端
1. 在 `web/` 目录添加前端代码
2. 配置端口转发

### 部署
构建 Docker 镜像：
```bash
docker build -t my-gin-app .
docker run -p 8080:8080 my-gin-app
```

## 资源链接

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [VSCode DevContainer 文档](https://code.visualstudio.com/docs/devcontainers/containers)
- [Go 官方文档](https://golang.org/doc/)