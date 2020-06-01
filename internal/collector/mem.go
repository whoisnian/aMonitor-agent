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

type memInfo struct {
	RAMTotal    int64
	RAMCached   int64
	RAMUsed     int64
	RAMFree     int64
	RAMAvail    int64
	RAMUsedPCT  int64
	SwapTotal   int64
	SwapUsed    int64
	SwapFree    int64
	SwapUsedPCT int64
}

// StartMEM 上报服务器RAM和swap状态
func StartMEM(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()
	fi, err := os.Open(meminfoFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var mem memInfo
	var value int64

	ticker := time.NewTicker(time.Duration(interval.MEM) * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			log.Println("Close memInfo collector.")
			return
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))

			pos := 0
			mem.RAMCached = 0
			for i := 0; i <= len(content); i++ {
				if i != len(content) && content[i] != '\n' {
					continue
				}

				arr := strings.Fields(content[pos:i])
				pos = i + 1
				if len(arr) < 2 {
					continue
				}

				util.StrToNumber(arr[1], &value)
				switch arr[0] {
				case "MemTotal:":
					mem.RAMTotal = value
				case "MemFree:":
					mem.RAMFree = value
				case "MemAvailable:":
					mem.RAMAvail = value
				case "Buffers:":
					mem.RAMCached += value
				case "Cached:":
					mem.RAMCached += value
				case "SReclaimable:":
					mem.RAMCached += value
				case "SwapTotal:":
					mem.SwapTotal = value
				case "SwapFree:":
					mem.SwapFree = value
				}
			}
			mem.RAMUsed = mem.RAMTotal - mem.RAMFree - mem.RAMCached
			mem.SwapUsed = mem.SwapTotal - mem.SwapFree
			if mem.RAMTotal <= 0 {
				mem.RAMUsedPCT = 0
			} else {
				mem.RAMUsedPCT = (mem.RAMTotal - mem.RAMAvail) * 10000 / mem.RAMTotal
			}
			if mem.SwapTotal <= 0 {
				mem.SwapUsedPCT = 0
			} else {
				mem.SwapUsedPCT = mem.SwapUsed * 10000 / mem.SwapTotal
			}

			select {
			case msgChan <- mem:
			case <-time.After(time.Second):
				log.Println("Timeout when send memInfo to msgChan.")
			}
		}
	}
}
