# builder 源镜像
FROM golang:1.12.4-alpine as builder

# 安装系统级依赖
RUN echo http://mirrors.aliyun.com/alpine/edge/main > /etc/apk/repositories \
&& echo http://mirrors.aliyun.com/alpine/edge/community >> /etc/apk/repositories \
&& apk update \
&& apk add --no-cache bash git gcc g++ openssl-dev librdkafka-dev pkgconf \
&& rm -rf /var/cache/apk/* \
&& git clone https://github.com/PharbersDeveloper/kafka-secrets /kafka-secrets

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
COPY --from=0 /kafka-secrets/snakeoil-ca-1.crt /kafka-secrets/
COPY --from=0 /kafka-secrets/kafkacat-ca1-signed.pem /kafka-secrets/
COPY --from=0 /kafka-secrets/kafkacat.client.key /kafka-secrets/
COPY deploy-config/kafka_config.json /resources/kafka_config.json

ENTRYPOINT ["/app/bp-jobs-observer"]
