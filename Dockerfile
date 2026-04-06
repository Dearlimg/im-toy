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

# 复制程序和静态文件
COPY --from=builder /app/im-chat .
COPY --from=builder /app/public ./public
COPY --from=builder /app/config/config.yaml ./config/config.yaml

EXPOSE 8080
CMD ["./im-chat"]