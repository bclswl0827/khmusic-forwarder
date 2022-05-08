package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var config mediaInfo

type mediaInfo struct {
	hlsLink    string
	hoursAvail int64
	m3u8Dir    string
	ffmpegPath string
}

// 获取直播流可用时长，并提前一小时
func validTime(uTime int64) int64 {
	// 获取本地 UTC 时间
	currentTime := time.Now().Local().UTC()
	// 解析传入的 Unix 时间戳，转为 UTC 时间
	futureTime := time.Unix(uTime, 0).Add(-8 * time.Hour).Local().UTC()
	validHours, _ := strconv.ParseInt(fmt.Sprintf("%.f", futureTime.Sub(currentTime).Hours()), 10, 64)
	return validHours - 1
}

// 获取 URL 参数资讯
func urlPraser(myUrl string, urlParam string) int64 {
	u, err := url.Parse(myUrl)
	if err != nil {
		panic(err)
	}
	p, _ := url.ParseQuery(u.RawQuery)
	// 转成 int64
	r, _ := strconv.ParseInt(p[urlParam][0], 10, 64)
	return r
}

// 解析 HTML 页面内部流媒体链接
func getLink() {
	log.Println("开始解析并获取流媒体地址")
    // 3 秒超时
	client := http.Client{Timeout: 3 * time.Second}
	for {
		res, err := client.Get("https://audio.voh.com.tw/KwongWah/m3u8.aspx")
		if err != nil {
			log.Println("因网络超时而重试中")
			continue
		}
		defer res.Body.Close()

		doc, _ := goquery.NewDocumentFromReader(res.Body)

		// 保存数据到全局变量
		doc.Find("source").Each(func(i int, s *goquery.Selection) {
			// 获取 HLS 地址
			config.hlsLink, _ = s.Attr("src")
			log.Println(config.hlsLink)
			// 获取过期日期
			config.hoursAvail = validTime(urlPraser(config.hlsLink, "expires"))
			log.Println("上述地址在", config.hoursAvail+1, "小时后过期")
		})
		break
	}
}

// 运行 FFmpeg
func ffmpeg(hlsLink string, hoursAvail int64, ffmpegPath string, m3u8Dir string) {
	for {
		// 启动 FFmpeg
		args := []string{"-y", "-nostats", "-nostdin", "-hide_banner",
			"-reconnect", "1", "-reconnect_at_eof", "1", "-reconnect_streamed", "1",
			"-reconnect_delay_max", "0", "-timeout", "2000000000", "-thread_queue_size", "5512",
			"-fflags", "+genpts", "-probesize", "10000000", "-analyzeduration", "15000000",
			"-i", hlsLink, "-c", "copy", "-segment_list_flags", "+live", "-hls_time", "4",
			"-hls_list_size", "6", "-hls_wrap", "10", "-segment_list_type", "m3u8",
			"-segment_time", "4", m3u8Dir + "/index.m3u8"}
		cmd := exec.Command(ffmpegPath, args...)
		err := cmd.Start()
		if err != nil {
			log.Fatalf("%s\n", err)
		}
		log.Println("已启动 FFmpeg 做流媒体转发")
		// 等待指定时长，然后结束 FFmpeg 进程
		time.Sleep(time.Duration(hoursAvail) * time.Hour)
		cmd.Process.Kill()
		cmd.Wait()
		// 获取新链接
		getLink()
		log.Println("FFmpeg 因流媒体链接变更而重启")
	}
}

// 判断文件夹是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func main() {
	// 指定命令行默认参数
	flag.StringVar(&config.ffmpegPath, "p", "/usr/bin/ffmpeg", "FFmpeg 绝对路径")
	flag.StringVar(&config.m3u8Dir, "o", "/www/khmusic", "TS 分片输出路径")
	flag.Parse()
	log.Println("FFmpeg 路径为", config.ffmpegPath)
	log.Println("HLS 流将会存放至", config.m3u8Dir)

	// 若输出文件夹不存在则创建
	dirExist, err := pathExists(config.m3u8Dir)
	if err != nil {
		panic(err)
	}
	if !dirExist {
		// 创建多级文件夹
		err := os.MkdirAll(config.m3u8Dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		log.Println("指定文件夹不存在，将会自动创建")
	}

	// 获取流媒体链接
	getLink()
	// 启动 FFmpeg
	ffmpeg(config.hlsLink, config.hoursAvail, config.ffmpegPath, config.m3u8Dir)
}
