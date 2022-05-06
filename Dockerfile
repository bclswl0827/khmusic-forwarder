FROM alpine:latest as builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache --virtual .build-deps go \
    && env CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct go build -ldflags="-s -w" -o /khmusic

FROM alpine:latest

LABEL maintainer "Yuki Kikuchi <bclswl0827@yahoo.co.jp>"
COPY --from=builder /khmusic /usr/bin/khmusic

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache ffmpeg \
    && mkdir -p /www

VOLUME ["/www"]
CMD ["/usr/bin/khmusic"]
