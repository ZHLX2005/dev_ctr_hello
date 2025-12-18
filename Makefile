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
DOCKER_IMAGE=dev_ctr_hello
DOCKER_TAG=latest

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