package dao

import (
	"errors"
	"fmt"
	"fuckworld/lib/golog"
	"io"
	"io/fs"
	"net/http"
	"os"
)

const (
	imageDir = "data/img"
)

var (
	imgFileMap = make(map[string]*fs.FileInfo)
)

type ImageView struct {
	L *golog.Logger
}

func InitImageView() error {
	if err := os.MkdirAll(imageDir, 0777); err != nil {
		golog.Warn("[mkdir all fail] [err:%s] [dir:%s]", err, imageDir)
		return err
	}
	entrys, err := os.ReadDir(imageDir)
	if err != nil {
		golog.Warn("[read dir fail] [err:%s]", err)
		return err
	}
	golog.Info("[read dir] [len:%d]", len(entrys))

	for _, e := range entrys {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			golog.Info("[get file info fail] [err:%s]", err)
			return err
		}
		imgFileMap[info.Name()] = &info
	}
	golog.Info("[init image map] [len:%d]", len(imgFileMap))
	return nil
}

func (iv *ImageView) Save(name, url string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = iv.doSave(name, url)
		if err == nil {
			return nil
		}
		if err != nil {
			golog.Warn("[download image fail] [err:%s]", err)
		}
	}
	return err
}

func (iv *ImageView) doSave(name, url string) error {
	if _, ok := imgFileMap[name]; ok {
		iv.L.Info("[image exist] [name:%s]", name)
		return nil
	}
	// 发起 HTTP GET 请求
	response, err := http.Get(url)
	if err != nil {
		iv.L.Warn("[download image fail] [err:%s]", err)
		return err
	}
	if response.StatusCode != 200 {
		iv.L.Warn("[download image fail] [httpCode:%d]", response.StatusCode)
		return errors.New("download image fail")
	}
	defer response.Body.Close()

	// 创建本地文件来保存图像
	file, err := os.Create(fmt.Sprintf("%s/%s", imageDir, name))
	if err != nil {
		iv.L.Warn("[failed to create file] [err:%s]", err)
		return err
	}
	defer file.Close()

	// 保存文件
	_, err = io.Copy(file, response.Body)
	if err != nil {
		iv.L.Warn("[failed to copy image] [err:%s]", err)
		return err
	}

	// 缓存文件信息
	info, err := file.Stat()
	if err != nil {
		iv.L.Warn("[get file stat fail] [err:%s]", err)
		return err
	}
	imgFileMap[name] = &info
	return nil
}
