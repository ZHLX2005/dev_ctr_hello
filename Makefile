# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Binary name
BINARY_NAME=main
BINARY_UNIX=$(BINARY_NAME)_unix

# Build info
VERSION?=v1.0.0
BUILD_TIME=$(shell date +%Y-%m-%dT%T%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker
DOCKER_IMAGE=file-server
DOCKER_TAG=latest

# Keys and storage
KEYS_DIR=./keys
STORAGE_DIR=./storage

.PHONY: help
help: ## 显示帮助信息
	@echo "可用的命令："
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## 下载依赖
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: build
build: ## 构建应用
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) -v .

.PHONY: build-linux
build-linux: ## 构建 Linux 应用
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_UNIX) -v .

.PHONY: test
test: ## 运行测试
	$(GOTEST) -v ./...

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

.PHONY: test-bench
test-bench: ## 运行基准测试
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: run
run: ## 运行应用
	$(GOCMD) run main.go

.PHONY: run-dev
run-dev: ## 开发模式运行（使用 air 热重载）
	@if ! command -v air &> /dev/null; then \
		echo "正在安装 air..."; \
		$(GOINSTALL) github.com/cosmtrek/air@latest; \
	fi
	air

.PHONY: clean
clean: ## 清理构建文件
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)
	rm -f coverage.out
	rm -f coverage.html
	rm -rf tmp/

.PHONY: fmt
fmt: ## 格式化代码
	$(GOFMT) -s -w .

.PHONY: vet
vet: ## 检查代码
	$(GOCMD) vet ./...

.PHONY: lint
lint: ## 代码检查
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "正在安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	golangci-lint run

.PHONY: check
check: fmt vet lint ## 运行所有检查

.PHONY: install
install: ## 安装到本地
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .

.PHONY: dev-setup
dev-setup: ## 设置开发环境
	$(GOINSTALL) github.com/cosmtrek/air@latest
	$(GOINSTALL) github.com/swaggo/swag/cmd/swag@latest
	$(GOINSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## 推送 Docker 镜像
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: release
release: clean test build ## 发布构建
	@echo "准备发布版本: $(VERSION)"

# ===== 文件服务器相关命令 =====

.PHONY: keys-generate
keys-generate: ## 生成 RSA 密钥对
	@echo "生成 RSA 密钥对..."
	@if exist "scripts\generate_keys.ps1" ( \
		powershell -ExecutionPolicy Bypass -File scripts\generate_keys.ps1 \
	) else ( \
		bash scripts/generate_keys.sh \
	)

.PHONY: keys-setup
keys-setup: ## 设置密钥目录
	@mkdir -p $(KEYS_DIR) $(STORAGE_DIR)

.PHONY: setup
setup: keys-setup keys-generate ## 初始化文件服务器（生成密钥）
	@echo "文件服务器初始化完成!"
	@echo "公钥: $(KEYS_DIR)/public.pem"
	@echo "私钥: $(KEYS_DIR)/private.pem"
	@echo "注意: 请妥善保管私钥，不要提交到版本控制"

.PHONY: docker-up
docker-up: ## 启动 Docker Compose 服务
	docker-compose up -d

.PHONY: docker-down
docker-down: ## 停止 Docker Compose 服务
	docker-compose down

.PHONY: docker-logs
docker-logs: ## 查看 Docker Compose 日志
	docker-compose logs -f

.PHONY: docker-restart
docker-restart: docker-down docker-up ## 重启 Docker Compose 服务

.PHONY: server-start
server-start: setup docker-up ## 启动文件服务器（首次运行会生成密钥）
	@echo "文件服务器已启动!"
	@echo "访问地址: http://localhost:8080"
	@echo "健康检查: http://localhost:8080/health"

.PHONY: clean-all
clean-all: clean ## 清理所有文件（包括密钥和存储）
	@echo "清理密钥和存储目录..."
	@rm -rf $(KEYS_DIR) $(STORAGE_DIR)

.PHONY: init
init: ## 初始化项目
	@echo "初始化项目..."
	@if [ ! -f go.mod ]; then \
		$(GOMOD) init dev_ctr_hello; \
	fi
	$(MAKE) deps
	$(MAKE) dev-setup

# 默认目标
.DEFAULT_GOAL := help