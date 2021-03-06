package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
)

var config serverInfo

type serverInfo struct {
	dir  string
	port string
}

func main() {
	// 指定命令行默认参数
	flag.StringVar(&config.dir, "d", "/www", "静态文件路径")
	flag.StringVar(&config.port, "p", "8080", "HTTP 端口")
	flag.Parse()
	log.Println("静态文件路径为", config.dir)
	log.Println("HTTP 端口为", config.port)
	log.Println("HTTP 服务已经启动")
	// 设置 m3u8 的 MIME Type
	mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	log.Println(
		http.ListenAndServe(":"+config.port, http.FileServer(http.Dir(config.dir))),
	)
}
