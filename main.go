package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type BingImageGallery struct {
	Title string
	Data  BingImageGalleryData
}

type BingImageGalleryData struct {
	Images []BingImageItem
}

type BingImageItem struct {
	Caption     string
	Title       string
	Description string
	Date        string
	IsoDate     string
	ImageUrls   BingImageInfo
}

type BingImageInfo struct {
	Landscape BingImageUrl
}

type BingImageUrl struct {
	HighDef      string
	UltraHighDef string
	Wallpaper    string
}

type BingWallpaper struct {
	Title       string `json:"title"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Url         string `json:"url"`
}

// httpGet GET请求
func httpGet(path string) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// getBingWallpaperSourceData 获取bing首页壁纸原始数据
func getBingWallpaperSourceData(resourceDomain string) []BingWallpaper {
	var bingImages BingImageGallery
	var bingWallpaperList []BingWallpaper

	body, err := httpGet(resourceDomain + "/hp/api/v1/imagegallery?format=json")
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(body, &bingImages)
	if err != nil {
		fmt.Println("error:", err)
	}
	for _, v := range bingImages.Data.Images {
		bingWallpaperList = append(bingWallpaperList, BingWallpaper{
			Title:       v.Title,
			Caption:     v.Caption,
			Description: v.Description,
			Date:        v.IsoDate,
			Url:         resourceDomain + v.ImageUrls.Landscape.UltraHighDef,
		})
	}
	return bingWallpaperList
}

// saveBingWallpaperData 保存数据到本地
func saveBingWallpaperData(data []BingWallpaper, saveDir string) {
	byteBuf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(data)
	os.WriteFile(saveDir+"db.json", byteBuf.Bytes(), 0644)
}

// readLoactionData 读取当前本地数据
func readLoactionData(savePath string) []BingWallpaper {
	var bingWallpaperList []BingWallpaper
	byteData, err := os.ReadFile(savePath + "db.json")
	if err != nil {
		fmt.Println("error:", err)
	}
	err = json.Unmarshal(byteData, &bingWallpaperList)
	if err != nil {
		fmt.Println("error:", err)
	}
	return bingWallpaperList
}

// dataMerge 数据合并(新旧数据合并)
func dataMerge(newList []BingWallpaper, oldList []BingWallpaper) []BingWallpaper {
	var allList []BingWallpaper
	// 旧数据最近的一项
	lastItem := oldList[0]
	// 旧数据中最近一项在新数据中index
	lastIndexAtNewList := -1
	for i, v := range newList {
		if v.Date == lastItem.Date {
			lastIndexAtNewList = i
		}
	}

	// 旧数据中最近一项不在新数据中，直接合并
	if lastIndexAtNewList == -1 {
		allList = newList
	} else {
		allList = newList[0:lastIndexAtNewList]
	}
	allList = append(allList, oldList...)
	return allList
}

// parseFileName  通过文件路径解析文件名称
func parseFileName(resourceUrl string) string {
	u, err := url.Parse(resourceUrl)
	if err != nil {
		log.Fatalln(err)
	}
	return u.Query().Get("id")
}

// downloadImage 下载图片到本地
func downloadImage(done chan bool, imgUrl string, savePath string) {
	b, err := httpGet(imgUrl)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile(savePath+parseFileName(imgUrl), b, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	done <- true
}

// bingWallpaperBatchDownload 批量下载
func bingWallpaperBatchDownload(data []BingWallpaper, savePath string) {
	fmt.Println("download start ...")
	done := make(chan bool)
	for _, v := range data {
		go downloadImage(done, v.Url, savePath)
	}
	for range data {
		<-done
	}
	fmt.Println("download complete!")
}

func main() {
	dbPath := flag.String("f", "./", "data file directory")
	savePath := flag.String("s", "./", "file directory")
	flag.Parse()
	localBingWallpaperList := readLoactionData(*dbPath)
	bingWallpaperList := getBingWallpaperSourceData("https://cn.bing.com")
	saveBingWallpaperData(dataMerge(bingWallpaperList, localBingWallpaperList), *dbPath)
	bingWallpaperBatchDownload(bingWallpaperList, *savePath)
}
