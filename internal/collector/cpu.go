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

type cpuInfo struct {
	UsedPCT int64
}

type cpuData struct {
	user      uint64
	nice      uint64
	system    uint64
	idle      uint64
	iowait    uint64
	irq       uint64
	softirq   uint64
	steal     uint64
	guest     uint64
	guestnice uint64
	total     uint64
}

// StartCPU 上报服务器CPU状态
func StartCPU(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()
	fi, err := os.Open(statFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var cur, sav cpuData
	var cpu cpuInfo

	firstRun := true
	ticker := time.NewTicker(time.Duration(interval.CPU) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Close cpuInfo collector.")
			return
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))

			// 从文件第一行读取cpu整体数据
			pos := 0
			for pos = 0; pos < len(content); pos++ {
				if content[pos] == '\n' {
					break
				}
			}

			arr := strings.Fields(content[:pos])
			util.StrToNumber(arr[1], &cur.user)
			util.StrToNumber(arr[2], &cur.nice)
			util.StrToNumber(arr[3], &cur.system)
			util.StrToNumber(arr[4], &cur.idle)
			util.StrToNumber(arr[5], &cur.iowait)
			util.StrToNumber(arr[6], &cur.irq)
			util.StrToNumber(arr[7], &cur.softirq)
			util.StrToNumber(arr[8], &cur.steal)
			util.StrToNumber(arr[9], &cur.guest)
			util.StrToNumber(arr[10], &cur.guestnice)

			cur.total = cur.user + cur.nice + cur.system + cur.idle + cur.iowait + cur.irq + cur.softirq + cur.steal

			userPeriod := (cur.user - cur.guest) - (sav.user - sav.guest)
			nicePeriod := (cur.nice - cur.guestnice) - (sav.nice - sav.guestnice)
			systemAllPeriod := (cur.system + cur.irq + cur.softirq) - (sav.system + sav.irq + sav.softirq)
			stealPeriod := cur.steal - sav.steal
			guestPeriod := (cur.guest + cur.guestnice) - (sav.guest + sav.guestnice)
			totalPeriod := cur.total - sav.total

			if totalPeriod <= 0 {
				cpu.UsedPCT = 0
			} else {
				cpu.UsedPCT = int64((nicePeriod + userPeriod + systemAllPeriod +
					stealPeriod + guestPeriod) * 10000 / totalPeriod)
			}

			sav = cur

			if firstRun {
				firstRun = false
			} else {
				select {
				case msgChan <- cpu:
				case <-time.After(time.Second):
					log.Println("Timeout when send cpuInfo to msgChan.")
				}
			}
		}
	}
}
