package main

import (
	"encoding/json"
	"fuckworld/lib/dao"
	"fuckworld/lib/golog"
	"fuckworld/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	// "github.com/gocolly/colly/proxy"
)

func main() {
	Init()
	chapters()
}

func Init() {
	os.MkdirAll("logs", 0777)
	err := golog.SetFile("logs/fuck_spider.log")
	if err != nil {
		golog.Error("set golog.File error")
		return
	}
	golog.SetLevel(golog.LEVEL_INFO)
	golog.SetbackupCount(36) // log time to live: 2 days
	golog.EnableRotate(time.Hour)
	dao.Init()
}

func chapters() {
	host := "https://dontstarve.huijiwiki.com/"
	// 内容
	itemCollector := colly.NewCollector(colly.Async(false))
	itemCollector.SetRequestTimeout(time.Minute)
	extensions.RandomUserAgent(itemCollector)
	extensions.Referer(itemCollector)
	itemCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 20})
	itemCollector.OnError(func(resp *colly.Response, e error) {
		url := resp.Request.URL.String()
		golog.Warn("[visit item fail] [logid:%s] [err:%s] [url:%s]", resp.Ctx.Get("logid"), e, url)
	})
	itemCollector.OnHTML(`table[class="infobox"]`, func(ele *colly.HTMLElement) {
		log := golog.New()
		log.SetLogid(ele.Response.Ctx.Get("logid"))
		imageView := &dao.ImageView{L: log}
		itemView := &dao.ItemView{L: log}
		item := map[string]string{}

		ele.ForEach("tr", func(i int, p *colly.HTMLElement) {
			// 物品名称
			itemName(p, item, log)

			// 物品图片+图片名称
			itemImage(p, i, log, imageView, item)

			// 合成材料
			hecheng(p, log, imageView, item)
		})

		// return
		// 保存信息
		if err := itemView.Insert(item); err != nil {
			log.Warn("[insert item fail] [err:%s] [item:%+v]", err, item)
		}
	})

	// proxySwitcher, _ := proxy.RoundRobinProxySwitcher(
	// //"http://199.19.225.250:80",
	// )
	// 列表
	contentCollector := colly.NewCollector()
	// contentCollector.SetProxyFunc(proxySwitcher)
	contentCollector.SetRequestTimeout(time.Minute)
	extensions.RandomUserAgent(contentCollector)
	extensions.Referer(contentCollector)
	contentCollector.Limit(&colly.LimitRule{DomainGlob: "*"})
	contentCollector.OnError(func(resp *colly.Response, e error) {
		url := resp.Request.URL.String()
		golog.Warn("[visit item fail] [logid:%s] [err:%s] [url:%s]", resp.Ctx.Get("logid"), e, url)
	})
	contentCollector.OnHTML(`div[class="mw-category mw-category-columns"]`, func(ele *colly.HTMLElement) {
		ele.ForEach("a[href]", func(i int, p *colly.HTMLElement) {
			logid := utils.GenerateLogId()
			p.Request.Ctx.Put("logid", logid)
			p.Request.Ctx.Put("name", p.Text)
			ctx := colly.NewContext()
			ctx.Put("logid", logid)
			ctx.Put("name", p.Text)
			itemCollector.Request("GET", host+p.Attr("href"), nil, ctx, nil)
		})
	})

	contentCollector.Visit("https://dontstarve.huijiwiki.com/index.php?title=%E5%88%86%E7%B1%BB:%E7%89%A9%E5%93%81&pageuntil=jiang#mw-pages")
	contentCollector.Wait()
}

func hecheng(p *colly.HTMLElement, log *golog.Logger, imageView *dao.ImageView, item map[string]string) {
	// 合成材料
	if strings.TrimSpace(p.ChildText(`th[class="infobox-label"]`)) == "材料" {
		// 材料数量 按，或、分割
		countStr := p.ChildText(`td[class="infobox-data"]`)
		// 最后一个为合成必要条件时,获取的字符串是()
		hasMast := false
		if strings.Contains(countStr, "（") {
			hasMast = true
		}
		countStrArr := utils.SplitNumStr(countStr, "[，, 、\\s x × （ ）]+")
		var hechengs []*dao.HeCheng
		p.ForEach(`a`, func(i int, h *colly.HTMLElement) {
			nameCN := h.Attr(`title`)
			nameEN := h.ChildAttr(`img`, "alt")
			href := h.ChildAttr(`img`, "src")
			count := "1"
			if len(countStrArr) > i {
				count = countStrArr[i]
			} else {
				log.Debug("[hecheng] [no count] [count_str:%s] [count_len:%d] [name_cn:%s] [name_en:%s]", countStr, len(countStrArr), nameCN, nameEN)
			}
			// image alt可能为中文，此时取a标签中href
			// <a href="/wiki/%E6%96%87%E4%BB%B6:Papyrus.png" class="image" title="莎草纸">
			if utils.ContainsChineseCharacters(nameEN) {
				hrefA := h.Attr(`href`)
				if index := strings.Index(hrefA, ":"); index > -1 {
					nameEN = hrefA[index+1:]
				}
			}
			// a 标签中href也有可能为中文，此时取href中名称
			// src="https://huiji-thumb.huijistatic.com/dontstarve/uploads/thumb/e/e9/Saffron_Feather.png/32px-Saffron_Feather.png"
			if utils.ContainsChineseCharacters(nameEN) {
				nameEN = filepath.Base(href)
				strs := strings.Split(nameEN, "-")
				if len(strs) > 1 {
					nameEN = strs[1]
				}
			}
			log.Info("[hecheng] [name_en:%s] [name_cn:%s] [href:%s] [count:%s]", nameEN, nameCN, href, count)
			if href != "" {
				if err := imageView.Save(nameEN, href); err != nil {
					log.Warn("[save image fail] [err:%s] [name_cn:%s] [name_en:%s] [href:%s]", err, nameCN, nameEN, href)
					// return
				}
			}
			hecheng := &dao.HeCheng{NameEN: nameEN, NameCN: nameCN, Count: count}
			hechengs = append(hechengs, hecheng)
		})
		if hasMast {
			hechengs[len(hechengs)-1].Mast = true
		}
		data, err := json.Marshal(hechengs)
		if err != nil {
			log.Warn("[marshal hecheng fail] [err:%s] [hecheng:%+v]", err, hechengs)
			return
		}
		if len(hechengs) != len(countStrArr) {
			log.Warn("[hecheng count image not equles] [count_str:%s] [count_len:%d] [image_len:%d] [image_str:%s]", countStr, len(countStrArr), len(hechengs), string(data))
		}
		item["hecheng"] = string(data)
	}
}

// 物品图片+图片名称
func itemImage(p *colly.HTMLElement, i int, log *golog.Logger, imageView *dao.ImageView, item map[string]string) {
	imageName := p.ChildAttr(`div[class="inv_item inv_item_old"] a[class="image"] img`, "alt")
	imageURL := p.ChildAttr(`div[class="inv_item inv_item_old"] a[class="image"] img`, "src")

	// 一般第二个为物品图片，特殊格式没有div[class="inv_item inv_item_old"]
	if len(imageName) < 1 && i == 1 {
		imageName = p.ChildAttr(`a[class="image"] img`, "alt")
		imageURL = p.ChildAttr(`a[class="image"] img`, "src")
	}
	if len(imageName) > 0 {
		log.Info("[image_name:%s] [image_url:%s]", imageName, imageURL)
		if err := imageView.Save(imageName, imageURL); err != nil {
			log.Warn("[save image fail] [err:%s] [name:%s] [href:%s]", err, imageName, imageURL)

		}
		item["image_name"] = imageName
	}
}

// 物品名称
func itemName(p *colly.HTMLElement, item map[string]string, log *golog.Logger) {
	if p.Attr("class") == "infobox-title" {
		nameCN := p.ChildText(`big`)
		nameEN := p.ChildText(`small`)
		item["name_cn"] = nameCN
		item["name_en"] = nameEN
		log.Info("[%s:%s]", nameCN, nameEN)
	}
}
