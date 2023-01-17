package main

import (
	"log"
	"net/url"
	"os"
)

// ParseFileName  通过文件路径解析文件名称
func ParseFileName(resourceUrl string) string {
	u, err := url.Parse(resourceUrl)
	if err != nil {
		log.Fatalln(err)
	}
	return u.Query().Get("id")
}

// downloadImage 下载图片到本地
func DownloadImage(done chan bool, imgUrl string, savePath string) {
	b, err := httpGet(imgUrl)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile(savePath+ParseFileName(imgUrl), b, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	done <- true
}

// BingWallpaperBatchDownload 批量下载
func BingWallpaperBatchDownload(data []BingWallpaper, savePath string) {
	done := make(chan bool)
	for _, v := range data {
		go DownloadImage(done, v.Url, savePath)
	}
	for range data {
		<-done
	}
}
