# 使用官方 Go 镜像
FROM golang:1.25-bookworm

# 设置维护者信息
LABEL maintainer="your-email@example.com"
LABEL description="Custom Go Development Container with Gin Framework"

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV GIN_MODE=debug

# 创建 vscode 用户（如果不存在）
RUN if id "vscode" >/dev/null 2>&1; then \
        echo "User vscode already exists"; \
    else \
        useradd -m -s /bin/bash vscode && \
        usermod -aG sudo vscode && \
        echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers; \
    fi

# 安装系统依赖
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        # 基础工具
        curl \
        wget \
        git \
        vim \
        tree \
        htop \
        zip \
        unzip \
        # 构建工具
        build-essential \
        make \
        # 网络工具
        netcat-openbsd \
        # 清理缓存
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# 创建项目目录
WORKDIR /workspace

# 创建必要的目录结构
RUN mkdir -p /workspace/{cmd,internal,pkg,configs,docs,scripts,web,build}

# 设置目录权限
RUN chown -R vscode:vscode /workspace

# 暴露常用端口
EXPOSE 8080

# 切换到 vscode 用户
USER vscode

# 设置默认命令
CMD ["sleep", "infinity"]