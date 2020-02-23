package collector

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/whoisnian/aMonitor-agent/util"
)

type diskLoad struct {
	ReadPS    int64
	ReadSize  int64
	WritePS   int64
	WriteSize int64
}

type diskData struct {
	RD  uint64
	RDS uint64
	WR  uint64
	WRS uint64
}

// StartDisk 上报服务器的磁盘io状态
func StartDisk(msgChan chan interface{}) {
	fi, err := os.Open(diskstatsFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var cur, sav diskData
	var load diskLoad
	var value uint64

	ticker := time.NewTicker(time.Duration(interval.DISK) * time.Second)
	interrupt := make(chan os.Signal, 1)
	for {
		select {
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
				res := strings.Fields(content[pos:i])
				pos = i + 1
				if len(res) < 14 {
					continue
				}
				if res[1] != "0" {
					continue
				}

				util.StrToNumber(res[3], &value)
				cur.RD += value
				util.StrToNumber(res[5], &value)
				cur.RDS += value
				util.StrToNumber(res[7], &value)
				cur.WR += value
				util.StrToNumber(res[9], &value)
				cur.WRS += value
			}
			load.ReadPS = int64(cur.RD-sav.RD) / interval.DISK
			load.WritePS = int64(cur.WR-sav.WR) / interval.DISK
			// sector size: 512 bytes
			// sec * 512 (byte) / 1024 = sec / 2 (KB)
			load.ReadSize = int64(cur.RDS-sav.RDS) / 2
			load.WriteSize = int64(cur.WRS-sav.WRS) / 2
			sav = cur
			msgChan <- load
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
