package collector

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/whoisnian/aMonitor-agent/util"
	"golang.org/x/sys/unix"
)

type mountsInfo struct {
	Mounts []mountsData
}

type mountsData struct {
	DevName    string
	Point      string
	FsType     string
	TotalSize  int64
	FreeSize   int64
	AvailSize  int64
	UsedSizeP  int64
	TotalNodes int64
	FreeNodes  int64
	UsedNodesP int64
}

// StartMounts 上报服务器的挂载点磁盘用量
func StartMounts(msgChan chan interface{}) {
	fi, err := os.Open(mountsFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var mountsinfo mountsInfo
	var data mountsData

	content := string(util.ReadAll(fi))
	pos := 0
	for i := 0; i <= len(content); i++ {
		if i != len(content) && content[i] != '\n' {
			continue
		}

		res := strings.Fields(content[pos:i])
		pos = i + 1
		if len(res) < 3 {
			continue
		}

		if !filepath.IsAbs(res[0]) {
			continue
		}

		data.DevName = res[0]
		data.Point = res[1]
		data.FsType = res[2]
		mountsinfo.Mounts = append(mountsinfo.Mounts, data)
	}

	ticker := time.NewTicker(time.Duration(interval.MOUNTS) * time.Second)
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-ticker.C:
			for index := range mountsinfo.Mounts {
				var t unix.Statfs_t
				unix.Statfs(mountsinfo.Mounts[index].Point, &t)
				mountsinfo.Mounts[index].TotalSize = int64(t.Blocks>>10) * t.Bsize
				mountsinfo.Mounts[index].FreeSize = int64(t.Bfree>>10) * t.Bsize
				mountsinfo.Mounts[index].AvailSize = int64(t.Bavail>>10) * t.Bsize
				mountsinfo.Mounts[index].TotalNodes = int64(t.Files)
				mountsinfo.Mounts[index].FreeNodes = int64(t.Ffree)
				if t.Blocks == 0 {
					mountsinfo.Mounts[index].UsedSizeP = 0
				} else {
					mountsinfo.Mounts[index].UsedSizeP = int64((t.Blocks - t.Bfree) * 10000 / t.Blocks)
				}
				if t.Files == 0 {
					mountsinfo.Mounts[index].UsedNodesP = 0
				} else {
					mountsinfo.Mounts[index].UsedNodesP = int64((t.Files - t.Ffree) * 10000 / t.Files)
				}
			}
			msgChan <- mountsinfo
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
