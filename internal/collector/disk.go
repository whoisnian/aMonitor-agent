package collector

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/whoisnian/aMonitor-agent/internal/util"
)

type diskInfo struct {
	ReadPS    int64
	ReadSize  int64
	ReadRate  int64
	WritePS   int64
	WriteSize int64
	WriteRate int64
}

type diskData struct {
	RD  uint64
	RDS uint64
	WR  uint64
	WRS uint64
}

// StartDisk 上报服务器的磁盘io状态
func StartDisk(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()
	fi, err := os.Open(diskstatsFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var cur, sav diskData
	var disk diskInfo
	var value uint64

	firstRun := true
	ticker := time.NewTicker(time.Duration(interval.DISK) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Close diskInfo collector.")
			return
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))
			cur.RD, cur.WR, cur.RDS, cur.WRS = 0, 0, 0, 0
			pos := 0
			for i := 0; i <= len(content); i++ {
				if i != len(content) && content[i] != '\n' {
					continue
				}

				// [0]major [1]minor [2]device name
				// [3]reads completed [4]reads merged  [5]sectors read  [6]time spent reading
				// [7]writes completed  [8]writes merged  [9]sectors written [10]time spent writing
				// [11]I/Os currently in progress [12]time spent doing I/Os [13]weighted time spent doing I/Os
				// [14]discards completed [15]discards merged [16]sectors discarded [17]time spent discarding
				arr := strings.Fields(content[pos:i])
				pos = i + 1
				if len(arr) < 14 {
					continue
				}
				if arr[1] != "0" {
					continue
				}

				util.StrToNumber(arr[3], &value)
				cur.RD += value
				util.StrToNumber(arr[5], &value)
				cur.RDS += value
				util.StrToNumber(arr[7], &value)
				cur.WR += value
				util.StrToNumber(arr[9], &value)
				cur.WRS += value
			}
			disk.ReadPS = int64(cur.RD-sav.RD) / interval.DISK
			disk.WritePS = int64(cur.WR-sav.WR) / interval.DISK
			// sector size: 512 bytes
			// sec * 512 (byte) / 1024 = sec / 2 (KB)
			disk.ReadSize = int64(cur.RDS-sav.RDS) / 2
			disk.WriteSize = int64(cur.WRS-sav.WRS) / 2
			disk.ReadRate = int64(cur.RDS-sav.RDS) * 512 / interval.DISK
			disk.WriteRate = int64(cur.WRS-sav.WRS) * 512 / interval.DISK

			sav = cur

			if firstRun {
				firstRun = false
			} else {
				select {
				case msgChan <- disk:
				case <-time.After(time.Second):
					log.Println("Timeout when send diskInfo to msgChan.")
				}
			}
		}
	}
}
