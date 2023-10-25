package main

import (
	"fmt"
	"fuckworld/lib/dao"
	"fuckworld/lib/golog"
	"path/filepath"
	"regexp"
	"time"
)

func main() {
	base := filepath.Base("https://huiji-thumb.huijistatic.com/dontstarve/uploads/thumb/e/e9/Saffron_Feather.png/32px-Saffron_Feather.png")
	fmt.Println(base)
}

func containsChineseCharacters(s string) bool {
	// 正则表达式匹配中文字符
	re := regexp.MustCompile("[\\p{Han}]")
	return re.MatchString(s)
}

func Init() {
	err := golog.SetFile("logs/fuck.log")
	if err != nil {
		golog.Error("set golog.File error")
		return
	}
	golog.SetLevel(golog.LEVEL_INFO)
	golog.SetbackupCount(36) // log time to live: 2 days
	golog.EnableRotate(time.Hour)
	dao.Init()
}
