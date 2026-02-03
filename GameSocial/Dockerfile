# 构建阶段：使用Go 1.22-alpine，匹配go.mod的版本要求
FROM golang:1.22-alpine AS builder
# 设置工作目录
WORKDIR /app
# 复制go.mod/go.sum并下载依赖（利用Docker层缓存，避免重复下载）
COPY go.mod go.sum ./
RUN go mod download
# 复制所有项目代码
COPY . .
# 编译：关闭CGO，编译为linux/amd64架构的静态可执行文件，指向cmd/server入口
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

# 运行阶段：使用轻量alpine镜像，减小镜像体积
FROM alpine:3.18
# 安装时区依赖（解决Go程序时区问题，可选但推荐）
RUN apk add --no-cache tzdata
# 设置工作目录
WORKDIR /app
# 从构建阶段复制编译好的可执行文件
COPY --from=builder /app/main .
# 暴露云托管强制要求的8080端口
EXPOSE 8080
# 启动Go服务
CMD ["./main"]