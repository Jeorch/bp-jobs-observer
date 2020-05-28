# 源镜像
FROM golang:1.12.4-alpine

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

COPY kafka_config.json /resources/kafka_config.json

ENV GOPROXY="https://goproxy.cn"

# 使用脚本自动拉取最新release代码
COPY bootstrap.sh /bootstrap.sh
ENTRYPOINT ["sh", "/bootstrap.sh"]