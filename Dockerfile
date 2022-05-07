FROM alpine:latest as builder

COPY ./src /srv
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache --virtual .build-deps go \
    && cd /srv; env CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct go build -ldflags="-s -w" -o /usr/bin/khmusic

FROM alpine:latest

LABEL maintainer "Yuki Kikuchi <bclswl0827@yahoo.co.jp>"
COPY --from=builder /usr/bin/khmusic /usr/bin/khmusic
COPY entrypoint.sh /entrypoint.sh

ENV HTTP_ENABLED=false \
    HTTP_PORT=80
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache ffmpeg python3 \
    && mkdir -p /www

VOLUME ["/www"]
ENTRYPOINT ["/entrypoint.sh"]
