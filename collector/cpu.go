package collector

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/whoisnian/aMonitor-agent/util"
)

type cpuLoad struct {
	Avg int64
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
func StartCPU(msgChan chan interface{}) {
	fi, err := os.Open(statFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var cur, sav cpuData
	var load cpuLoad

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))
			fmt.Sscanf(content, "cpu%d%d%d%d%d%d%d%d%d%d",
				&cur.user, &cur.nice, &cur.system,
				&cur.idle, &cur.iowait, &cur.irq, &cur.softirq,
				&cur.steal, &cur.guest, &cur.guestnice)

			cur.total = cur.user + cur.nice + cur.system + cur.idle + cur.iowait + cur.irq + cur.softirq + cur.steal

			userPeriod := (cur.user - cur.guest) - (sav.user - sav.guest)
			nicePeriod := (cur.nice - cur.guestnice) - (sav.nice - sav.guestnice)
			systemAllPeriod := (cur.system + cur.irq + cur.softirq) - (sav.system + sav.irq + sav.softirq)
			stealPeriod := cur.steal - sav.steal
			guestPeriod := (cur.guest + cur.guestnice) - (sav.guest + sav.guestnice)
			totalPeriod := cur.total - sav.total

			load.Avg = int64((nicePeriod + userPeriod + systemAllPeriod +
				stealPeriod + guestPeriod) * 10000 / totalPeriod)

			if sav.total > 0 {
				msgChan <- load
			}

			sav = cur
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
