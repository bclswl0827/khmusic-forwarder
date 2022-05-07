#!/bin/sh

if [ $HTTP_ENABLED ]; then
    echo "HTTP server enabled"
    # 在 Heroku 等平台上，存在值为随机数的 PORT 变量
    if [ $PORT ]; then
        /usr/bin/python3 -m http.server -d /www $PORT --bind :: &
    else
        # 如果不存在 PORT 变量，则检查用户指定的 HTTP_PORT 变量是否存在
        if [ $HTTP_PORT ]; then
            /usr/bin/python3 -m http.server -d /www $HTTP_PORT --bind :: &
        else
            echo "No HTTP port specified"
            exit 1
        fi
    fi
fi

# 启动转发进程
/usr/bin/khmusic
