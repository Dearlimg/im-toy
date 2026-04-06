# 编译阶段
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

RUN apk add --no-cache gcc musl-dev ca-certificates tzdata
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o im-chat ./cmd/main.go

# 运行阶段
FROM alpine:3.19
WORKDIR /app

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

# 复制程序
COPY --from=builder /app/im-chat .

# 【关键】手动创建 config 文件夹，再复制文件！
RUN mkdir -p /app/config
COPY --from=builder /app/config/config.yaml /app/config/config.yaml

# 权限
RUN chmod -R 755 /app

# 不用非root用户，避免权限坑（先跑起来再说）
# USER appuser

EXPOSE 8080
CMD ["./im-chat"]