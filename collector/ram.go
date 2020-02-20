package collector

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/whoisnian/aMonitor-agent/util"
)

type memInfo struct {
	MTotal  int64
	MCached int64
	MUsed   int64
	MFree   int64
	MAvail  int64
	MAvg    int64
	STotal  int64
	SUsed   int64
	SFree   int64
	SAvg    int64
}

// StartRAM 上报服务器RAM和swap状态
func StartRAM(msgChan chan interface{}) {
	fi, err := os.Open(meminfoFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var mem memInfo
	var value int64

	ticker := time.NewTicker(time.Duration(interval.RAM) * time.Second)
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-ticker.C:
			content := string(util.SeekAndReadAll(fi))

			pos := 0
			mem.MCached = 0
			for i := 0; i <= len(content); i++ {
				if i != len(content) && content[i] != '\n' {
					continue
				}

				res := strings.Fields(content[pos:i])
				if len(res) >= 2 {
					util.StrToNumber(res[1], &value)
				}
				pos = i + 1

				switch res[0] {
				case "MemTotal:":
					mem.MTotal = value
				case "MemFree:":
					mem.MFree = value
				case "MemAvailable:":
					mem.MAvail = value
				case "Buffers:":
					mem.MCached += value
				case "Cached:":
					mem.MCached += value
				case "SReclaimable:":
					mem.MCached += value
				case "SwapTotal:":
					mem.STotal = value
				case "SwapFree:":
					mem.SFree = value
				}
			}
			mem.MUsed = mem.MTotal - mem.MFree - mem.MCached
			mem.MAvg = (mem.MFree + mem.MCached) * 10000 / mem.MTotal
			mem.SUsed = mem.STotal - mem.SFree
			mem.SAvg = mem.SUsed * 10000 / mem.STotal

			msgChan <- mem
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
