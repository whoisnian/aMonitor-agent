package util

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// FileExists 文件存在且非目录
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists 目录存在
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// ReadAll 读取文件全部内容（用于单次读取）
func ReadAll(fi *os.File) []byte {
	result, err := ioutil.ReadAll(fi)
	if err != nil {
		log.Panicln(err)
	}
	return result
}

// SeekAndReadAll seek至文件开头再读取文件全部内容（用于重复多次读取）
func SeekAndReadAll(fi *os.File) []byte {
	fi.Seek(0, 0)

	var result []byte
	buf := make([]byte, 4096)
	for {
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			log.Panicln(err)
		}
		if 0 == n {
			break
		}
		result = append(result, buf[:n]...)
	}
	return result
}

// StrToNumber 字符串转数字(参数为字符串和数值变量指针)
func StrToNumber(str string, numP interface{}) {
	str = strings.TrimSpace(str)

	var err error
	switch p := numP.(type) {
	case *int64:
		*p, err = strconv.ParseInt(str, 10, 64)
	case *uint64:
		*p, err = strconv.ParseUint(str, 10, 64)
	case *float64:
		*p, err = strconv.ParseFloat(str, 64)
	}
	if err != nil {
		log.Panicln(err)
	}
}
