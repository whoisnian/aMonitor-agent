package collector

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/whoisnian/aMonitor-agent/util"
)

type loadInfo struct {
	Avg1  float64
	Avg5  float64
	Avg15 float64
}

// StartLoad 上报服务器平均负载
func StartLoad(msgChan chan interface{}) {
	fi, err := os.Open(loadavgFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var load loadInfo

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))

			res := strings.Fields(content)
			load.Avg1, _ = strconv.ParseFloat(res[0], 64)
			load.Avg5, _ = strconv.ParseFloat(res[1], 64)
			load.Avg15, _ = strconv.ParseFloat(res[2], 64)

			msgChan <- load
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
