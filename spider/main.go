package main

import (
	"encoding/json"
	"fmt"
	"fuckworld/conf"
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

// 刷数据
func main1() {
	Init()
	itemArr := dao.ItemArr()
	// Food & Gardening Filter.png

	GameNameCN := make(map[string]struct{})
	GameImgNameEN := make(map[string]struct{})
	ZhiZuoNameCN := make(map[string]struct{})
	ZhiZuoImgNameEN := make(map[string]struct{})
	ZhiZuoImgNameCNEmpty := make(map[string]struct{})
	for _, item := range itemArr {
		fenlei, ok := item["联机分类"]
		if !ok || len(fenlei) < 1 {
			continue
		}
		strs := strings.Split(fenlei, ",")
		for _, str := range strs {
			_, ok1 := conf.HERO_IMG_MAP[str]
			_, ok2 := conf.TYPE_IMG_MAP[str]
			if !ok1 && !ok2 {
				ZhiZuoImgNameEN[str] = struct{}{}
			}
		}
	}

	golog.Info("GameNameCN: %+v", GameNameCN)
	golog.Info("GameImgNameEN: %+v", GameImgNameEN)
	golog.Info("ZhiZuoNameCN: %+v", ZhiZuoNameCN)
	golog.Info("ZhiZuoImgNameEN: %+v", utils.MapToSlice[string, struct{}](ZhiZuoImgNameEN))
	golog.Info("ZhiZuoImgNameCNEmpty: %+v", utils.MapToSlice[string, struct{}](ZhiZuoImgNameCNEmpty))
}

// 刷数据
func main2() {
	Init()
	itemArr := dao.ItemArr()
	// Food & Gardening Filter.png

	NameMap := make(map[string]string)
	NameCN2ENMap := make(map[string]string)
	ZhiZuoImgNameCNEmpty := make(map[string]struct{})
	ZhiZuoImgNameENEmpty := make(map[string]struct{})
	for _, item := range itemArr {
		fenlei, ok := item["diao_luo_zi"]
		if !ok || len(fenlei) < 1 {
			continue
		}
		var hecheng []dao.BasePair
		err := json.Unmarshal([]byte(fenlei), &hecheng)
		if err != nil {
			golog.Error("[unmarshal fail] [err:%s]", err)
			continue
		}
		for _, zhizuo := range hecheng {
			// zhizuo.NameCN = strings.TrimLeft(zhizuo.NameCN, "文件:")
			if len(zhizuo.NameCN) > 0 && len(zhizuo.ImageNameEN) > 0 {
				NameMap[zhizuo.ImageNameEN] = zhizuo.NameCN
				NameCN2ENMap[zhizuo.NameCN] = zhizuo.ImageNameEN
			}
			if len(zhizuo.NameCN) < 1 {
				// golog.Info("[empty ZhiZuoImgNameEN] [%+v]", item)
				ZhiZuoImgNameCNEmpty[zhizuo.ImageNameEN] = struct{}{}
			}
			if len(zhizuo.ImageNameEN) < 1 {
				// golog.Info("[empty ZhiZuoImgNameEN] [%+v]", item)
				ZhiZuoImgNameENEmpty[zhizuo.NameCN] = struct{}{}
			}

			// if "Don't Starve icon.png" == zhizuo.JieSuoImgNameEN {
			// 	golog.Info("[Wigfrid Portrait.png] [%+v]", item)
			// }
		}
	}

	en2cn := "\n"
	for en, _ := range ZhiZuoImgNameCNEmpty {
		en2cn += fmt.Sprintf("\"%s\":\"%s\",\n", en, NameMap[en])
	}
	cn2en := "\n"
	for cn, _ := range ZhiZuoImgNameENEmpty {
		cn2en += fmt.Sprintf("\"%s\":\"%s\",\n", cn, NameCN2ENMap[cn])
	}
	fmt.Println(en2cn)
	fmt.Println("=========")
	fmt.Println(cn2en)

	// golog.Info("ZhiZuoImgNameCNEmpty: %+v", utils.MapToSlice[string, struct{}](ZhiZuoImgNameCNEmpty))
	// golog.Info("ZhiZuoImgNameENEmpty: %+v", utils.MapToSlice[string, struct{}](ZhiZuoImgNameENEmpty))
}

// 爬数据
func main() {
	Init()

	// 基础物品
	itemListURL := []string{
		`https://dontstarve.huijiwiki.com/index.php?title=%E5%88%86%E7%B1%BB:%E7%89%A9%E5%93%81&pageuntil=jiang#mw-pages`,
		`https://dontstarve.huijiwiki.com/index.php?title=%E5%88%86%E7%B1%BB:%E7%89%A9%E5%93%81&pagefrom=jiang#mw-pages`,
		`https://dontstarve.huijiwiki.com/index.php?title=%E5%88%86%E7%B1%BB:%E7%89%A9%E5%93%81&pagefrom=xi+gua+bing+gun#mw-pages`,
	}
	for _, u := range itemListURL {
		Items(u)
	}

	// 补充食物表
	// AppendFoodItem("https://dontstarve.huijiwiki.com/wiki/%E7%83%B9%E9%A5%AA")
}

// 补充食物信息
func AppendFoodItem(url string) {
	host := "https://dontstarve.huijiwiki.com/"

	// 合成示例
	itemCollector := colly.NewCollector(colly.Async(false))
	itemCollector.SetRequestTimeout(time.Minute)
	extensions.RandomUserAgent(itemCollector)
	extensions.Referer(itemCollector)
	itemCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 20})
	itemCollector.OnError(func(resp *colly.Response, e error) {
		url := resp.Request.URL.String()
		golog.Warn("[visit item fail] [logid:%s] [err:%s] [url:%s]", resp.Ctx.Get("logid"), e, url)
	})
	itemCollector.OnHTML(`div[style="clear:left"]`, func(h *colly.HTMLElement) {
		log := golog.New()
		log.SetLogid(h.Response.Ctx.Get("logid"))
		itemNameCN := h.Response.Ctx.Get("name")
		item := h.Response.Ctx.GetAny("item").(*map[string]string)
		imageView := &dao.ImageView{L: log}
		var zhizuoArr [][]dao.BasePair
		zhizuo := (*item)["shiwu_zhizuo"]
		json.Unmarshal([]byte(zhizuo), &zhizuoArr)
		var zhizuoOne []dao.BasePair
		h.ForEach(`div[class="inv_back inv_back_old"]`, func(i int, h *colly.HTMLElement) {
			nameCN := h.ChildAttr(`div[class="inv_item inv_item_old"] a`, `title`)
			imageNameEN := h.ChildAttr(`div[class="inv_item inv_item_old"] img`, `alt`)
			imageURL := h.ChildAttr(`div[class="inv_item inv_item_old"] img`, `src`)
			if imageURL != "" {
				if err := imageView.Save(imageNameEN, imageURL); err != nil {
					log.Warn("[save image fail] [err:%s] [item_name:%s] [name_cn:%s] [img_name_en:%s] [img_url:%s]", err, itemNameCN, nameCN, imageNameEN, imageURL)
					// return
				}
			}
			it := dao.BasePair{
				NameCN:      nameCN,
				ImageNameEN: imageNameEN,
			}
			zhizuoOne = append(zhizuoOne, it)
		})
		zhizuoArr = append(zhizuoArr, zhizuoOne)
		data, err := json.Marshal(zhizuoArr)
		if err != nil {
			log.Warn("[marshal fail] [err:%s] [it:%+v]", err, zhizuoArr)
			return
		}
		(*item)["shiwu_zhizuo"] = string(data)
	})

	// 列表
	listCollector := itemCollector.Clone()
	listCollector.OnHTML(`table[class="wikitable sortable"]`, func(h *colly.HTMLElement) {
		h.ForEach("tbody tr", func(i int, tr *colly.HTMLElement) {
			log := golog.New()
			log.SetLogid(utils.GenerateLogId())
			imageView := &dao.ImageView{L: log}
			itemView := &dao.ItemView{L: log}
			var item map[string]string
			tr.ForEach("td", func(i int, td *colly.HTMLElement) {
				// 食物图片+名称
				if i == 0 {
					nameCN := td.ChildAttr(`a`, "title")
					imageNameEN := td.ChildAttr(`img`, "alt")
					imageURL := td.ChildAttr(`img`, "src")
					if imageURL != "" {
						if err := imageView.Save(imageNameEN, imageURL); err != nil {
							log.Warn("[save image fail] [err:%s] [name_cn:%s] [img_name_en:%s] [img_url:%s]", err, nameCN, imageNameEN, imageURL)
							// return
						}
					}
					item = itemView.Select("name_cn", nameCN)
					log.Info("[select] [name_cn:%s] [res:%+v]", nameCN, item)
					if item == nil {
						item = make(map[string]string)
						item["name_cn"] = nameCN
						item["image_name"] = imageNameEN
					}

					// 跳转合成列表
					ctx := colly.NewContext()
					ctx.Put("logid", log.Logid())
					ctx.Put("name", nameCN)
					ctx.Put("item", &item)
					itemCollector.Request("GET", host+td.ChildAttr(`a`, "href"), nil, ctx, nil)
				}

				// 支持的版本
				if i == 2 || i == 3 {
					nameCN := td.ChildAttr(`a`, "title")
					if len(nameCN) < 1 {
						return
					}
					imageNameEN := td.ChildAttr(`img`, "alt")
					imageURL := td.ChildAttr(`img`, "src")
					if imageURL != "" {
						if err := imageView.Save(imageNameEN, imageURL); err != nil {
							log.Warn("[save image fail] [err:%s] [name_cn:%s] [img_name_en:%s] [img_url:%s]", err, nameCN, imageNameEN, imageURL)
							// return
						}
					}
					it := dao.BasePair{
						NameCN:      nameCN,
						ImageNameEN: imageNameEN,
					}
					data, err := json.Marshal(it)
					if err != nil {
						log.Warn("[marshal fail] [err:%s] [it:%+v]", err, it)
						return
					}
					item[fmt.Sprintf("version%d", i-1)] = string(data)
				}

				// 烹饪时间
				if i == 8 {
					item["pengrenshijian"] = td.Text
				}

				// 优先级
				if i == 9 {
					item["youxianji"] = td.Text
				}

				// 属性要求
				if i == 10 {
					item["shuxingyaoqiu"] = td.Text
				}

				// 指定
				if i == 11 {
				}
			})
			if len(item) < 1 {
				return
			}
			err := itemView.Update(item, "name_cn", item["name_cn"])
			if err != nil {
				log.Warn("[update fail] [name_cn:%s] [item:%+v]", item["name_cn"], item)
			}
			log.Info("[update item] [name_cn:%s] [item:%+v]", item["name_cn"], item)
		})
	})
	listCollector.Visit(url)
	listCollector.Wait()
}

func Items(url string) {
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

			// 单机分类
			fenlei("制作栏", "单机分类", p, log, imageView, item, func(imgNameEN string) string {
				name, ok := conf.TYPE_SINGLE_TO_OLINE[imgNameEN]
				if ok {
					return name
				}
				return imgNameEN
			})

			// 联机分类
			fenlei("制作分类", "联机分类", p, log, imageView, item, nil)

			// 解锁条件
			fenlei("解锁", "解锁条件", p, log, imageView, item, nil)

			// 功能
			if str, match := simpleText(p, "功能"); match {
				item["功能"] = str
			}

			// 伤害
			if str, match := simpleText(p, "伤害"); match {
				item["伤害"] = str
			}

			// 耐久
			if str, match := simpleText(p, "耐久"); match {
				item["耐久"] = str
			}

			// 代码
			if str, match := simpleText(p, "代码"); match {
				item["代码"] = strings.Trim(str, `"`)
			}

			// 掉落自
			diaoluozi(p, log, imageView, item)

			// 食物属性
			simpleFoodAttr(p, log, imageView, item, "生命值", "生命值")
			simpleFoodAttr(p, log, imageView, item, "饥饿值", "饥饿值")
			simpleFoodAttr(p, log, imageView, item, "理智值", "理智值")

			//食物类型
			if str, match := simpleText(p, "食物类型"); match {
				item["食物类型"] = str
			}

			//烹饪优先级
			if str, match := simpleText(p, "优先级"); match {
				item["烹饪优先级"] = str
			}

			//烹饪时间
			if str, match := simpleText(p, "时间"); match {
				item["烹饪时间"] = str
			}

			//腐烂时间
			if str, match := simpleText(p, "腐烂时间"); match {
				item["腐烂时间"] = str
			}

			//堆叠上限
			if str, match := simpleText(p, "堆叠上限"); match {
				item["堆叠上限"] = str
			}

			// 制作范例

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

	contentCollector.Visit(url)
	contentCollector.Wait()
}

// 食物属性
func simpleFoodAttr(p *colly.HTMLElement, log *golog.Logger, imageView *dao.ImageView, item map[string]string, attrName, columnName string) {
	if strings.TrimSpace(p.ChildAttr(`th[class="infobox-label"] a`, "title")) == attrName {
		nameCN := attrName
		imageNameEN := p.ChildAttr(`img`, "alt")
		imageURL := p.ChildAttr(`img`, "src")
		value := strings.TrimSpace(p.ChildText(`td[class="infobox-data"]`))
		if imageURL != "" {
			if err := imageView.Save(imageNameEN, imageURL); err != nil {
				log.Warn("[save image fail] [err:%s] [name_cn:%s] [img_name_en:%s] [img_url:%s]", err, nameCN, imageNameEN, imageURL)
				// return
			}
		}
		// attr := dao.FoodAttr{
		// 	NameCN:      nameCN,
		// 	ImageNameEN: imageNameEN,
		// 	Value:       value,
		// }
		// data, err := json.Marshal(attr)
		// if err != nil {
		// 	log.Warn("[marshal diaoluozi fail] [err:%s] [jiesuo:%+v]", err, attr)
		// 	return
		// }
		item[columnName] = value
	}
}

// 如果th text为thText，获取td内容
func simpleText(p *colly.HTMLElement, thText string) (string, bool) {
	if strings.TrimSpace(p.ChildText(`th[class="infobox-label"]`)) == thText {
		return strings.TrimSpace(p.ChildText(`td[class="infobox-data"]`)), true
	}
	return "", false
}

// 掉落自
func diaoluozi(p *colly.HTMLElement, log *golog.Logger, imageView *dao.ImageView, item map[string]string) {
	if strings.TrimSpace(p.ChildText(`th[class="infobox-label"]`)) == "掉落自" {
		var diaoLuoZiS []dao.BasePair
		p.ForEach("a", func(i int, p *colly.HTMLElement) {
			// 制作分类
			jieSuoNameCN := p.Attr("title")
			jieSuoImgNameEN := p.ChildAttr(`img`, "alt")
			jieSuoImgURL := p.ChildAttr(`img`, "src")
			if jieSuoImgURL != "" {
				if err := imageView.Save(jieSuoImgNameEN, jieSuoImgURL); err != nil {
					log.Warn("[save image fail] [err:%s] [diaoluozi_name_cn:%s] [diaoluozi_img_name_en:%s] [diaoluozi_img_url:%s]", err, jieSuoNameCN, jieSuoImgNameEN, jieSuoImgURL)
					// return
				}
			}
			if _, ok := conf.DIAO_LUO_ERR[jieSuoImgNameEN]; ok {
				return
			}
			if len(jieSuoNameCN) < 1 {
				jieSuoNameCN = conf.TOTLE[jieSuoImgNameEN]
			}
			diaoLuoZi := dao.BasePair{
				NameCN:      jieSuoNameCN,
				ImageNameEN: jieSuoImgNameEN,
			}
			diaoLuoZiS = append(diaoLuoZiS, diaoLuoZi)
		})

		data, err := json.Marshal(diaoLuoZiS)
		if err != nil {
			log.Warn("[marshal diaoluozi fail] [err:%s] [jiesuo:%+v]", err, diaoLuoZiS)
			return
		}
		item["掉落自"] = string(data)
	}
}

// 解锁
func jiesuo(p *colly.HTMLElement, log *golog.Logger, imageView *dao.ImageView, item map[string]string) {
	if strings.TrimSpace(p.ChildText(`th[class="infobox-label"]`)) == "解锁" {
		// 制作分类
		jieSuoNameCN := p.ChildAttr(`td[class="infobox-data"] a`, "title")
		jieSuoImgNameEN := p.ChildAttr(`td[class="infobox-data"] img`, "alt")
		jieSuoImgURL := p.ChildAttr(`td[class="infobox-data"] img`, "src")
		if jieSuoImgURL != "" {
			if err := imageView.Save(jieSuoImgNameEN, jieSuoImgURL); err != nil {
				log.Warn("[save image fail] [err:%s] [jiesuo_name_cn:%s] [jiesuo_img_name_en:%s] [jiesuo_img_url:%s]", err, jieSuoNameCN, jieSuoImgNameEN, jieSuoImgURL)
				// return
			}
		}
		jiesuo := dao.JieSuo{
			JieSuoNameCN:    jieSuoNameCN,
			JieSuoImgNameEN: jieSuoImgNameEN,
		}
		data, err := json.Marshal(jiesuo)
		if err != nil {
			log.Warn("[marshal jiesuo fail] [err:%s] [jiesuo:%+v]", err, jiesuo)
			return
		}
		item["jie_suo"] = string(data)
	}
}

// 制作分类
func fenlei(altName, saveName string, p *colly.HTMLElement, log *golog.Logger, imageView *dao.ImageView, item map[string]string, onvert func(imgNameEN string) string) {
	if strings.TrimSpace(p.ChildText(`th[class="infobox-label"]`)) == altName {
		var basePairs []dao.BasePair
		p.ForEach(`td[class="infobox-data"] img`, func(i int, h *colly.HTMLElement) {
			imgNameEN := strings.TrimSpace(strings.Replace(h.Attr("alt"), " &", "", -1))
			imgURL := h.Attr("src")
			if imgURL != "" {
				if err := imageView.Save(imgNameEN, imgURL); err != nil {
					log.Warn("[save image fail] [err:%s] [img_name_en:%s] [img_url:%s]", err, imgNameEN, imgURL)
					// return
				}
			}
			if onvert != nil {
				imgNameEN = onvert(imgNameEN)
			}
			pair := dao.BasePair{
				NameCN:      conf.TOTLE[imgNameEN],
				ImageNameEN: imgNameEN,
			}
			basePairs = append(basePairs, pair)
		})
		if len(basePairs) < 1 {
			return
		}
		data, err := json.Marshal(basePairs)
		if err != nil {
			log.Warn("[marshal fail] [err:%s] [base_pairs:%+v]", err, basePairs)
			return
		}
		item[saveName] = string(data)
	}
}

// 制作栏
func zhizuolan(p *colly.HTMLElement, log *golog.Logger, imageView *dao.ImageView, item map[string]string) {
	if strings.TrimSpace(p.ChildText(`th[class="infobox-label"]`)) == "制作栏" {
		// 支持的游戏版本
		gameNameCN := p.ChildAttr(`th[class="infobox-label"] a`, "title")
		gameImgNameEN := p.ChildAttr(`th[class="infobox-label"] img`, "alt")
		gameImgURL := p.ChildAttr(`th[class="infobox-label"] img`, "src")
		if gameImgURL != "" {
			if err := imageView.Save(gameImgNameEN, gameImgURL); err != nil {
				log.Warn("[save image fail] [err:%s] [game_name_cn:%s] [game_img_name_en:%s] [game_img_url:%s]", err, gameNameCN, gameImgNameEN, gameImgURL)
				// return
			}
		}

		// 制作栏
		zhiZuoNameCN := p.ChildAttr(`td[class="infobox-data"] a`, "title")
		zhiZuoImgNameEN := p.ChildAttr(`td[class="infobox-data"] img`, "alt")
		zhiZuoImgURL := p.ChildAttr(`td[class="infobox-data"] img`, "src")
		if zhiZuoImgURL != "" {
			if err := imageView.Save(zhiZuoImgNameEN, zhiZuoImgURL); err != nil {
				log.Warn("[save image fail] [err:%s] [zhizuo_name_cn:%s] [zhizuo_img_name_en:%s] [zhizuo_img_url:%s]", err, zhiZuoNameCN, zhiZuoImgNameEN, zhiZuoImgURL)
				// return
			}
		}
		zhizuo := dao.ZhiZuoLan{
			GameNameCN:      gameNameCN,
			GameImgNameEN:   gameImgNameEN,
			ZhiZuoNameCN:    zhiZuoNameCN,
			ZhiZuoImgNameEN: zhiZuoImgNameEN,
		}
		data, err := json.Marshal(zhizuo)
		if err != nil {
			log.Warn("[marshal zhizuo fail] [err:%s] [zhizuolan:%+v]", err, zhizuo)
			return
		}
		item["zhi_zuo_lan"] = string(data)
	}
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
		countStrArr := utils.SplitNumStr(countStr, "[，, 、\\s x × （ ）/]+")
		var hechengs []*dao.ItemFormula
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
			hecheng := &dao.ItemFormula{BasePair: dao.BasePair{NameCN: nameCN, ImageNameEN: nameEN}, Count: count}
			hechengs = append(hechengs, hecheng)
		})
		if len(hechengs) < 1 {
			return
		}
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
		item["合成表"] = string(data)
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
		item["图片名称"] = imageName
	}
}

// 物品名称
func itemName(p *colly.HTMLElement, item map[string]string, log *golog.Logger) {
	if p.Attr("class") == "infobox-title" {
		nameCN := p.ChildText(`big`)
		nameEN := p.ChildText(`small`)
		item["中文名称"] = nameCN
		item["英文名称"] = nameEN
		log.Info("[%s:%s]", nameCN, nameEN)
	}
}
