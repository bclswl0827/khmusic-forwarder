#!/bin/sh

# 根据条件判断是否启用 HTTP 服务
if [ $HTTP_ENABLED ]; then
    echo "$(date '+%Y/%m/%d %H:%M:%S') HTTP 服务已被启用"
    # 在 Heroku 等平台上，存在值为随机数的 PORT 变量
    if [ $PORT ]; then
        /usr/bin/python3 -m http.server -d /www $PORT --bind :: &
    else
        # 如果不存在 PORT 变量，则检查用户指定的 HTTP_PORT 变量是否存在
        if [ $HTTP_PORT ]; then
            /usr/bin/python3 -m http.server -d /www $HTTP_PORT --bind :: &
        else
            echo "$(date '+%Y/%m/%d %H:%M:%S') 没有为 HTTP 服务指定端口"
            exit 1
        fi
    fi
fi

# 根据条件判断是否一并转发汉声电台
if [ $VOH_ENABLED ]; then
    echo "$(date '+%Y/%m/%d %H:%M:%S') 汉声电台将被一并转发"
    /usr/bin/voh &
fi

# 启动转发进程
/usr/bin/khmusic
