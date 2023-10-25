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

// 合成表
type HeCheng struct {
	NameEN string
	NameCN string
	Count  string // 数量
	Mast   bool // 是必要条件，不是合成材料
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
	return utils.WriteExcelWithTitle(itemPath, header, itemArrayToSave, nil, false)
}
