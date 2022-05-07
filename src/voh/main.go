package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var config mediaInfo

type mediaInfo struct {
	hlsLink    [2]string
	hoursAvail [2]int64
	m3u8Dir    [2]string
	ffmpegPath string
}

var stream streamInfo

type streamInfo struct {
	Url               string `json:"Url"`
	Token             string `json:"token"`
	Expires           string `json:"expires"`
	EndTime           string `json:"EndTime"`
	Subject           string `json:"Subject"`
	AMFM_PlayTimeWeek string `json:"AMFM_PlayTimeWeek"`
	PlayChannel       int    `json:"PlayChannel"`
}

// 获取直播流可用时长，并提前一小时
func validTime(uTime int64) int64 {
	currentTime := time.Now().UTC()
	futureTime := time.Unix(uTime, 0).UTC()
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
	// 依次解析 HLS 地址
	for i := 0; i < len(config.m3u8Dir); i++ {
		for {
			urlValues := url.Values{}
			urlValues.Add("Type", "VideoDisplay")
			urlValues.Add("PlayChannel", strconv.Itoa(i+1)) // 1 是 FM，2 是 AM
			res, err := client.PostForm(
				"https://audio.voh.com.tw/API/ugC_ProgramHandle.ashx",
				urlValues,
			)
			if err != nil {
				log.Println("因网络超时而重试中")
				continue
			}
			defer res.Body.Close()
			body, _ := ioutil.ReadAll(res.Body)
			json.Unmarshal([]byte(string(body)), &stream)
			// 保存数据到全局变量
			config.hlsLink[i] = stream.Url + "?token=" + stream.Token + "&expires=" + stream.Expires
			log.Println(config.hlsLink[i])
			// 获取过期日期
			config.hoursAvail[i] = validTime(urlPraser(config.hlsLink[i], "expires"))
			log.Println("上述地址在", config.hoursAvail[i]+1, "小时后过期")
			break
		}
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
	flag.StringVar(&config.m3u8Dir[0], "f", "/www/voh_fm", "汉声 FM 台 TS 分片输出路径")
	flag.StringVar(&config.m3u8Dir[1], "a", "/www/voh_am", "汉声 AM 台 TS 分片输出路径")
	flag.Parse()
	log.Println("FFmpeg 路径为", config.ffmpegPath)
	log.Println("汉声 FM 台 TS 分片输出路径", config.m3u8Dir[0])
	log.Println("汉声 AM 台 TS 分片输出路径", config.m3u8Dir[1])

	// 若输出文件夹不存在则创建
	for i := 0; i < len(config.m3u8Dir); i++ {
		dirExist, err := pathExists(config.m3u8Dir[i])
		if err != nil {
			panic(err)
		}
		if !dirExist {
			// 创建多级文件夹
			err := os.MkdirAll(config.m3u8Dir[i], os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
	}

	// 获取流媒体链接
	getLink()
	// 以协程启动 FFmpeg
	for i := 0; i < len(config.m3u8Dir); i++ {
		if i != 0 {
			ffmpeg(config.hlsLink[i], config.hoursAvail[i], config.ffmpegPath, config.m3u8Dir[i])
		} else {
			go ffmpeg(config.hlsLink[i], config.hoursAvail[i], config.ffmpegPath, config.m3u8Dir[i])
		}
	}
}
