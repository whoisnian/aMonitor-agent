package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type fileMDInfo struct {
	Path  string
	Event string
}

// aMonitor-agent插件消息监听地址
var addr string = "127.0.0.1:8001"

// 目标文件列表
var fileList []string

func init() {
	if envAddr, ok := unix.Getenv("ADDR"); ok {
		addr = envAddr
	}
	log.Println(addr)

	// 按':'分割出多个文件路径
	if envFileList, ok := unix.Getenv("FILELIST"); ok {
		for _, file := range strings.FieldsFunc(envFileList, func(c rune) bool { return c == ':' }) {
			fileList = append(fileList, path.Clean(file))
		}
	}
	log.Println(fileList)

	if len(fileList) < 1 {
		log.Panicln("no target file.")
	}
}

// 监控文件自身的修改移动删除
func addFileWatch(fd int, pathname string) int {
	wd, err := unix.InotifyAddWatch(fd, pathname, unix.IN_CLOSE_WRITE|unix.IN_MOVE_SELF|unix.IN_DELETE|unix.IN_DELETE_SELF)
	if err != nil {
		log.Println("ERR: addFileWatch", fd, pathname, err)
	}
	return wd
}

// 监控文件父级目录的创建事件，当发现目标文件被新建时不再监控旧文件，改为监控新文件
func addDirWatch(fd int, pathname string) int {
	wd, err := unix.InotifyAddWatch(fd, pathname, unix.IN_CREATE)
	if err != nil {
		log.Println("ERR: addDirWatch", fd, pathname, err)
	}
	return wd
}

// 发送事件
func sendInfo(info fileMDInfo) {
	log.Println(info)
	content, err := json.Marshal(info)
	if err != nil {
		log.Println(err)
	}

	resp, err := http.Post("http://"+addr+"?category=fileMDInfo", "application/json", bytes.NewBuffer(content))
	if err == nil {
		defer resp.Body.Close()
	}
}

func main() {
	fd, err := unix.InotifyInit()
	if err != nil {
		log.Panicln(err)
	}
	defer unix.Close(fd)

	// 对目标文件及其父级目录添加监视器
	wdNameMap := make(map[int]string) // 接收到file事件时从watchdesc还原文件或目录名
	nameWdMap := make(map[string]int) // 接收到dir事件时从完整文件名找到原watchdesc
	for _, file := range fileList {
		wd := addFileWatch(fd, file)
		wdNameMap[wd] = file
		nameWdMap[file] = wd

		wd = addDirWatch(fd, path.Dir(file))
		wdNameMap[wd] = path.Dir(file)
	}

	var fileMD fileMDInfo
	var buf [4096]byte
	var offset uint32
	var knownEvent bool

	for {
		n, err := unix.Read(fd, buf[:])
		if err != nil {
			log.Panicln(err)
		}

		offset = 0
		for int(offset) <= n-unix.SizeofInotifyEvent {
			event := (*unix.InotifyEvent)(unsafe.Pointer(&buf[offset]))

			var exist bool
			if fileMD.Path, exist = wdNameMap[int(event.Wd)]; exist {
				if event.Len > 0 {
					bytes := (*[unix.PathMax]byte)(unsafe.Pointer(&buf[offset+unix.SizeofInotifyEvent]))
					fileMD.Path = path.Join(fileMD.Path, strings.TrimRight(string(bytes[:event.Len-1]), "\000"))
				}

				if wd, ok := nameWdMap[fileMD.Path]; ok {
					knownEvent = true
					if unix.IN_ISDIR != event.Mask&unix.IN_ISDIR &&
						unix.IN_CREATE == event.Mask&unix.IN_CREATE {
						// 删除旧文件的监视器
						unix.InotifyRmWatch(fd, uint32(wd))
						delete(wdNameMap, wd)
						delete(nameWdMap, fileMD.Path)
						// 更新为新文件的监视器
						wd = addFileWatch(fd, fileMD.Path)
						wdNameMap[wd] = fileMD.Path
						nameWdMap[fileMD.Path] = wd
						fileMD.Event = "create"
					} else if unix.IN_CLOSE_WRITE == event.Mask&unix.IN_CLOSE_WRITE {
						fileMD.Event = "modify"
					} else if unix.IN_DELETE == event.Mask&unix.IN_DELETE ||
						unix.IN_DELETE_SELF == event.Mask&unix.IN_DELETE_SELF {
						fileMD.Event = "delete"
					} else if unix.IN_MOVE_SELF == event.Mask&unix.IN_MOVE_SELF {
						fileMD.Event = "move"
					} else {
						knownEvent = false
					}

					if knownEvent {
						sendInfo(fileMD)
					}
				}
			}

			offset += unix.SizeofInotifyEvent + event.Len
		}
	}
}
