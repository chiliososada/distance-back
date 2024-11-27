FROM golang:1.21-alpine

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev

# 安装开发工具
RUN go install github.com/cosmtrek/air@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# air 配置文件
COPY .air.toml .

# 启动开发服务器
CMD ["air", "-c", ".air.toml"]