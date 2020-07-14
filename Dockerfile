# builder 源镜像
FROM golang:1.12.4-alpine as builder

# 安装系统级依赖
RUN echo http://mirrors.aliyun.com/alpine/edge/main > /etc/apk/repositories \
&& echo http://mirrors.aliyun.com/alpine/edge/community >> /etc/apk/repositories \
&& apk update \
&& apk add --no-cache bash git gcc g++ openssl-dev librdkafka-dev pkgconf \
&& rm -rf /var/cache/apk/*

ENV GOPROXY="https://goproxy.cn"

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build

# prod 源镜像
FROM alpine:latest as prod

# 安装 主要 依赖
RUN echo http://mirrors.aliyun.com/alpine/edge/main > /etc/apk/repositories \
&& echo http://mirrors.aliyun.com/alpine/edge/community >> /etc/apk/repositories \
&& apk update \
&& apk add --no-cache bash git gcc g++ openssl-dev librdkafka-dev pkgconf \
&& apk add --no-cache tzdata \
&& ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo "Asia/Shanghai" > /etc/timezone \
&& apk del tzdata \
&& rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=0 /app/bp-jobs-observer .
COPY deploy-config/kafka_config.json /resources/kafka_config.json

ENTRYPOINT ["/app/bp-jobs-observer"]

# docker build -t '444603803904.dkr.ecr.cn-northwest-1.amazonaws.com.cn/bp-jobs-observer:0.0.15' .
