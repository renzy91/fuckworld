package utils

import (
	"fmt"
	"fuckworld/lib/golog"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	r  = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu = sync.RWMutex{}
)

type Context struct {
	L *golog.Logger
}

func TrimRightStr(s, substr string) string {
	index := strings.LastIndex(s, substr)
	// 不存在或不在s末尾
	if index < 0 || index+len(substr) != len(s) {
		return s
	}
	return s[:index]
}

// GenerateLogId 64bits: logId(0) + 35 bits(ms offset of this year) + 1 bit(0:not vip) + 27bits(random number)
func GenerateLogId() string {
	date := time.Now()
	base := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
	offset := time.Since(base).Milliseconds()
	mu.Lock()
	defer mu.Unlock()
	rand_num := r.Int63n(2<<27 - 1)
	logid := offset<<28 | rand_num
	return fmt.Sprintf("%d", logid)
}

func ArrayChunk[T any](s []T, size int) [][]T {
	if size < 1 {
		return [][]T{}
	}
	length := len(s)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var res [][]T
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		res = append(res, s[i*size:end])
		i++
	}
	return res
}

func StructFieldsToStringSlice(s interface{}) []string {
	var result []string
	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		result = append(result, fieldName)
	}
	return result
}

// map转struct，s为struct指针
func Map2Struct(m map[string]string, s interface{}) error {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i).Name
		ty := t.Field(i).Type.Kind()
		switch ty {
		case reflect.String:
			v.Field(i).SetString(m[field])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intV, err := strconv.ParseInt(m[field], 10, 64)
			if err != nil {
				return err
			}
			v.Field(i).SetInt(intV)
		}
	}
	return nil
}

// struct转map
// obj是struct,不能传指针
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

// 判断字符串中是否有中文字符
func ContainsChineseCharacters(s string) bool {
	// 正则表达式匹配中文字符
	re := regexp.MustCompile("[\\p{Han}]")
	return re.MatchString(s)
}

func IsNumeric(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

// 根据seps将str进行分割
// seps：分割符正则，如[，、\\s x]+ 按照, \ 空格 x分割
func Split(str, seps string) []string {
	re := regexp.MustCompile("[，、]")
	return re.Split(str, -1)
}

// 根据seps分割str，返回数字arr
// seps：分割符正则，如[，、\\s x]+ 按照, \ 空格 x分割
func SplitNumStr(str, reg string) []string {
	re := regexp.MustCompile(reg)
	strArr := re.Split(str, -1)
	var res []string
	for _, s := range strArr {
		if IsNumeric(s) {
			res = append(res, s)
		}
	}
	return res
}