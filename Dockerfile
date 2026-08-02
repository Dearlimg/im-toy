# syntax=docker/dockerfile:1
# 编译阶段
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 先下载依赖（利用 Docker 层缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译（去掉 -a 避免全量重编，使用 BuildKit 缓存加速二次构建）
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -ldflags '-w -s' -o im-chat ./cmd/main.go

# 运行阶段
FROM alpine:3.19
WORKDIR /app

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

# 复制程序和静态文件（配置由 docker-compose volumes 挂载，不在镜像内）
COPY --from=builder /app/im-chat .
COPY --from=builder /app/public ./public

EXPOSE 8080
CMD ["./im-chat"]
