package utils

import (
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func PageOffset(page, pagesize int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * pagesize
}

func GenerateRandomString(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 定义包含所有可能字符的字符串
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 生成随机字符串
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

func GenerateDateFilePath(len int, ext string) string {
	rs := GenerateRandomString(len) + strconv.FormatInt(time.Now().UnixMilli(), 10) + ext
	dir := time.Now().Format("2006-01-02")
	path := dir + "/" + rs
	return path
}

// 生成指定位数的随机数字字符串
func GenerateRandomNumber(length int) string {
	result := ""
	m := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}
	for i := 0; i < length; i++ {
		k := rand.Intn(len(m))
		result += m[k]
	}

	return result
}

func String(v string) *string {
	return &v
}

func Int(v int) *int {
	return &v
}

func Int32(v int32) *int32 {
	return &v
}

func Int64(v int64) *int64 {
	return &v
}

func Time(v time.Time) *time.Time {
	return &v
}

func GetFuncName(dest interface{}) string {
	pointer := reflect.ValueOf(dest).Pointer()
	func_name := runtime.FuncForPC(pointer).Name()
	//去除后缀-fm
	if strings.HasSuffix(func_name, "-fm") {
		func_name = strings.TrimRight(func_name, "-fm")
	}
	parts := strings.Split(func_name, "/")
	func_name = parts[len(parts)-1]
	return func_name
}
