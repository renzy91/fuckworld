package dao

import (
	"fuckworld/lib/golog"
	"fuckworld/utils"
	"os"
	"path/filepath"
)

const (
	itemPath = "data/db/item.xlsx"
)

var (
	header          []string
	headerMap       = make(map[string]struct{})
	itemArray       []map[string]string
	itemArrayToSave []map[string]interface{}
	// 记录唯一索引，用于去重
	uniqIndexNameEN    = make(map[string]struct{})
	uniqIndexNameENStr = "name_en"
)

type ItemView struct {
	L *golog.Logger
}

// 基础
type BasePair struct {
	NameCN      string
	ImageNameEN string
}

// 物品合成表
type ItemFormula struct {
	BasePair
	Count string // 数量
	Mast  bool   // 是必要条件，不是合成材料
}

// 合成表
type HeCheng struct {
	NameEN string
	NameCN string
	Count  string // 数量
	Mast   bool   // 是必要条件，不是合成材料
}

// 制作栏
type ZhiZuoLan struct {
	GameNameCN      string
	GameImgNameEN   string
	ZhiZuoNameCN    string
	ZhiZuoImgNameEN string
}

// 解锁
type JieSuo struct {
	JieSuoNameCN    string
	JieSuoImgNameEN string
}

// 食物属性
type FoodAttr struct {
	NameCN      string
	ImageNameEN string
	Value       string
}

func InitItemView() error {
	if err := os.MkdirAll(filepath.Dir(itemPath), 0777); err != nil {
		golog.Warn("[mkdir all fail] [err:%s] [dir:%s]", err, filepath.Dir(itemPath))
		return err
	}
	title, data, err := utils.ReadExcelWithTitle(itemPath, nil)
	if err != nil || len(data) == 0 {
		golog.Warn("[read item data fail] [path:%s] [err:%s] [len:%d]", itemPath, err, len(data))
		return nil
	}
	header = title
	itemArray = data

	// headerMap
	headerMap = make(map[string]struct{})
	for _, h := range header {
		headerMap[h] = struct{}{}
	}

	// itemArrayToSave
	for _, item := range itemArray {
		m := make(map[string]interface{})
		for k, v := range item {
			m[k] = v
		}
		itemArrayToSave = append(itemArrayToSave, m)
		// 记录唯一索引，用于去重
		indexNameEN := item[uniqIndexNameENStr]
		if len(indexNameEN) > 0 {
			uniqIndexNameEN[indexNameEN] = struct{}{}
		}
	}

	golog.Info("[read data] [item_len:%d] [data_len:%d]", len(itemArray), len(data))
	return nil
}

func ItemArr() []map[string]string {
	return itemArray
}

func (iv *ItemView) Select(field, value string) map[string]string {
	_, res := iv.doSelect(field, value)
	return res
}

func (iv *ItemView) doSelect(field, value string) (index int, res map[string]string) {
	if _, ok := headerMap[field]; !ok {
		return -1, nil
	}
	for i, item := range itemArray {
		if item[field] == value {
			return i, item
		}
	}
	return -1, nil
}

func (iv *ItemView) Update(it map[string]string, conditionKey, conditionValue string) error {
	index, res := iv.doSelect(conditionKey, conditionValue)
	if index > -1 {
		iv.doDelete(index, it)
		for k, v := range it {
			res[k] = v
		}
	} else {
		res = it
	}
	return iv.Insert(res)
}

func (iv *ItemView) doDelete(i int, it map[string]string) {
	indexNameEN, ok := it[uniqIndexNameENStr]
	if ok {
		delete(uniqIndexNameEN, indexNameEN)
	}
	if i < 0 || i > len(itemArray)-1 {
		return
	}
	itemArray[i] = itemArray[len(itemArray)-1]
	itemArray = itemArray[:len(itemArray)-1]
	itemArrayToSave[i] = itemArrayToSave[len(itemArrayToSave)-1]
	itemArrayToSave = itemArrayToSave[:len(itemArrayToSave)-1]
}

func (iv *ItemView) Insert(it map[string]string) error {
	// 数据已存在
	indexNameEN, ok := it[uniqIndexNameENStr]
	if ok {
		if _, ok := uniqIndexNameEN[indexNameEN]; ok {
			iv.L.Info("[item exist] [%s:%s]", uniqIndexNameENStr, indexNameEN)
			return nil
		}
	}

	itemArray = append(itemArray, it)

	m := make(map[string]interface{})
	for k, v := range it {
		if _, ok := headerMap[k]; !ok {
			headerMap[k] = struct{}{}
			header = append(header, k)
		}
		m[k] = v
	}
	itemArrayToSave = append(itemArrayToSave, m)
	return iv.save()
}

func (iv *ItemView) save() error {
	// iv.L.Info("[save item] [len:%d] [1:%+v] [2:%+v]", len(itemArrayToSave), itemArrayToSave[0], itemArray[0])
	return utils.WriteExcelWithTitle(itemPath, header, itemArrayToSave, nil, false)
}
