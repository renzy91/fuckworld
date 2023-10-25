package utils

import (
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"
)

const (
	// 默认sheet名称前缀
	DEFAULT_SHEET_NAME_PREFIX = "Sheet"

	// 每个sheet存储最大行数(不包含title)
	MAX_ROW_PER_SHEET = 30000
)

// sheet生成算法，sheetIndex从0开始
type NextSheetName func(sheetIndex int) string

// 读取带有表头的excel,默认每个sheet第一行为表头
// path: excel文件路径,test/a.xlsx
// tableHeader: 表头, tableHeader=nil，每个sheet第一行做为表头
// return：最后一个sheet表头、excel数据
func ReadExcelWithTitle(path string, tableHeader []string) ([]string, []map[string]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}

	// 读sheet
	sheets := f.GetSheetList()
	if len(sheets) < 1 {
		return nil, nil, nil
	}

	//
	var data []map[string]string
	header := tableHeader
	for _, sheet := range sheets {
		// 获取整个sheel
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, nil, err
		}

		// 遍历每行，第一行做为tableHeader
		header = tableHeader
		for i, row := range rows {
			if i == 0 && len(tableHeader) == 0 {
				for _, v := range row {
					header = append(header, v)
				}
			} else if i > 0 {
				d := make(map[string]string)
				for j, v := range row {
					d[header[j]] = v
				}
				data = append(data, d)
			}
		}
	}
	return header, data, nil
}

// 写excel
// path: excel文件路径,test/a.xlsx
// tableHeader: 表头
// data: 数据
// nsn: sheet名称生成算法，不传默认使用defaultNextSheetName(), Sheet1 ~ Sheetn
// delDefaultSheet: 删除默认sheet页(sheet1)，需要保证传入的nsn方法不生成Sheet1。nsn=nil时此参数无效
// saveHeader: 是否保存标题列，saveHeader=true，tableHeader会被保存到每个table第一行
func WriteExcelWithTitle(path string, tableHeader []string, data []map[string]interface{}, nsn NextSheetName, delDefaultSheet bool) error {
	if len(data) < 1 {
		return errors.New("empty data")
	}
	f := excelize.NewFile()
	defer f.Close()

	// sheet生成算法
	if nsn == nil {
		nsn = defaultNextSheetName
		delDefaultSheet = false
	}

	// 生成表头
	xAxis := genXAxis(len(tableHeader))

	// 按sheet拆分数据
	dataArr := ArrayChunk[map[string]interface{}](data, MAX_ROW_PER_SHEET)
	for sheetIndex, rows := range dataArr {
		// 创建sheet
		sheetName := nsn(sheetIndex)
		_, err := f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		rowIndex := 1
		// 写入表头
		for i, h := range tableHeader {
			if err := f.SetCellValue(sheetName, fmt.Sprintf("%s%d", xAxis[i], rowIndex), h); err != nil {
				return err
			}
		}
		rowIndex++

		// 写入数据
		for _, row := range rows {
			for i, x := range xAxis {
				if err := f.SetCellValue(sheetName, fmt.Sprintf("%s%d", x, rowIndex), row[tableHeader[i]]); err != nil {
					return err
				}
			}
			rowIndex++
		}
	}

	// 删除默认sheet
	if delDefaultSheet {
		if err := f.DeleteSheet("Sheet1"); err != nil {
			return err
		}
	}

	// 写文件
	return f.SaveAs(path)
}

// 默认sheet名称生成算法
func defaultNextSheetName(sheetIndex int) string {
	return fmt.Sprintf("%s%d", DEFAULT_SHEET_NAME_PREFIX, sheetIndex+1)
}

// 生成excel横轴名称
func genXAxis(size int) []string {
	xAxis := make([]string, size)
	for i := 1; i <= size; i++ {
		xAxis[i-1] = numToXAxis(i)
	}
	return xAxis
}

// 数字转换为excel的横轴
func numToXAxis(size int) string {
	a := size
	xAxis := ""
	for size > 0 {
		a = (size - 1) / 26
		b := (size - 1) % 26
		xAxis = string(byte(b+65)) + xAxis
		size = a
	}
	return xAxis
}
