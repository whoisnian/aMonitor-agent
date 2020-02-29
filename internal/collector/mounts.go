package collector

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/whoisnian/aMonitor-agent/internal/util"
	"golang.org/x/sys/unix"
)

type mountsInfo struct {
	Mounts []mountsData
}

type mountsData struct {
	DevName      string
	Point        string
	FsType       string
	TotalSize    int64
	FreeSize     int64
	AvailSize    int64
	UsedSizePCT  int64
	TotalNodes   int64
	FreeNodes    int64
	UsedNodesPCT int64
}

// StartMounts 上报服务器的挂载点磁盘用量
func StartMounts(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()
	fi, err := os.Open(mountsFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var mounts mountsInfo
	var data mountsData

	content := string(util.ReadAll(fi))
	pos := 0
	for i := 0; i <= len(content); i++ {
		if i != len(content) && content[i] != '\n' {
			continue
		}

		arr := strings.Fields(content[pos:i])
		pos = i + 1
		if len(arr) < 3 {
			continue
		}

		if !filepath.IsAbs(arr[0]) {
			continue
		}

		data.DevName = arr[0]
		data.Point = arr[1]
		data.FsType = arr[2]
		mounts.Mounts = append(mounts.Mounts, data)
	}

	ticker := time.NewTicker(time.Duration(interval.MOUNTS) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Close mountsInfo collector.")
			return
		case <-ticker.C:
			for index := range mounts.Mounts {
				var t unix.Statfs_t
				unix.Statfs(mounts.Mounts[index].Point, &t)
				mounts.Mounts[index].TotalSize = int64(t.Blocks>>10) * t.Bsize
				mounts.Mounts[index].FreeSize = int64(t.Bfree>>10) * t.Bsize
				mounts.Mounts[index].AvailSize = int64(t.Bavail>>10) * t.Bsize
				mounts.Mounts[index].TotalNodes = int64(t.Files)
				mounts.Mounts[index].FreeNodes = int64(t.Ffree)
				if t.Blocks == 0 {
					mounts.Mounts[index].UsedSizePCT = 0
				} else {
					mounts.Mounts[index].UsedSizePCT = int64((t.Blocks - t.Bfree) * 10000 / t.Blocks)
				}
				if t.Files == 0 {
					mounts.Mounts[index].UsedNodesPCT = 0
				} else {
					mounts.Mounts[index].UsedNodesPCT = int64((t.Files - t.Ffree) * 10000 / t.Files)
				}
			}
			select {
			case msgChan <- mounts:
			case <-time.After(time.Second):
				log.Println("Timeout when send mountsInfo to msgChan.")
			}
		}
	}
}
