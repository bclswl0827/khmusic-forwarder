FROM alpine:latest as builder

COPY ./src /srv
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache --virtual .build-deps go \
    && cd /srv/khmusic; env CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct go build -ldflags="-s -w" -o /usr/local/bin/khmusic \
    && cd /srv/voh; env CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct go build -ldflags="-s -w" -o /usr/local/bin/voh

FROM alpine:latest

LABEL maintainer "Yuki Kikuchi <bclswl0827@yahoo.co.jp>"
COPY --from=builder /usr/local/bin /usr/local/bin
COPY entrypoint.sh /entrypoint.sh

ENV HTTP_ENABLED=false \
    HTTP_PORT=80 \
    VOH_ENABLED=true
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache ffmpeg python3 \
    && mkdir -p /www \
    && chmod 755 /entrypoint.sh

VOLUME ["/www"]
ENTRYPOINT ["/entrypoint.sh"]
