# 构建阶段：使用 Go 1.22-alpine，匹配 go.mod 版本要求
FROM golang:1.22-alpine AS builder
#scp -r ./ root@47.107.174.153:/root/app/
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o gamesocial ./cmd/server

# 运行阶段：使用轻量 alpine 镜像
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/gamesocial ./gamesocial

ENV SERVER_PORT=8080
ENV MEDIA_DIR=/app/data/media

RUN mkdir -p /app/data/media

EXPOSE 8080

CMD ["./gamesocial"]
