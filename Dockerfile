# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置必要的环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 设置工作目录
WORKDIR /build

# 安装基础工具
RUN apk add --no-cache git make

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 运行阶段
FROM alpine:3.18

# 安装基础工具和证书
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/build/distance-back .
# 复制配置文件
COPY --from=builder /build/config/config.yaml ./config/
COPY --from=builder /build/config/config.prod.yaml ./config/

# 创建必要的目录
RUN mkdir -p /app/logs && \
    chmod +x /app/distance-back

# 暴露端口
EXPOSE 80

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/health || exit 1

# 启动应用
CMD ["./distance-back", "-env", "production"]