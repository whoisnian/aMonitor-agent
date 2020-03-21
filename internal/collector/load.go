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

type loadInfo struct {
	Avg1  int64
	Avg5  int64
	Avg15 int64
}

// StartLoad 上报服务器平均负载
func StartLoad(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()
	fi, err := os.Open(loadavgFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var load loadInfo
	var value float64

	ticker := time.NewTicker(time.Duration(interval.LOAD) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Close loadInfo collector.")
			return
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))

			arr := strings.Fields(content)
			util.StrToNumber(arr[0], &value)
			load.Avg1 = int64(value * 100)
			util.StrToNumber(arr[1], &value)
			load.Avg5 = int64(value * 100)
			util.StrToNumber(arr[2], &value)
			load.Avg15 = int64(value * 100)

			select {
			case msgChan <- load:
			case <-time.After(time.Second):
				log.Println("Timeout when send loadInfo to msgChan.")
			}
		}
	}
}
