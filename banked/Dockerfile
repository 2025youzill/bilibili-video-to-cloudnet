# 阶段1: 构建环境
FROM golang:1.24-alpine AS builder
WORKDIR /banked

# 安装依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev git make curl tar xz

COPY ./ ./
RUN make setup-linux-ffmpeg
RUN GOPROXY=https://goproxy.cn go mod download
RUN GOPROXY=https://goproxy.cn CGO_ENABLED=0 GOOS=linux \
    go build -v -ldflags="-s -w" -o main ./

# 阶段2: 生产镜像
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates wget && rm -rf /var/lib/apt/lists/*


WORKDIR /banked

COPY --from=builder /banked/main .
COPY --from=builder /banked/tool/ffmpeg/ffmpeg /banked/tool/ffmpeg/ffmpeg
COPY --from=builder /banked/cloudnet/data/cookie.json /banked/cloudnet/data/cookie.json

EXPOSE 8080
CMD ["./main"]