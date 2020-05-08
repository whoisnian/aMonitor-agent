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
)

type netInfo struct {
	Rrate    int64
	Rsum     int64
	Rpackets int64
	Trate    int64
	Tsum     int64
	Tpackets int64
}

type netData struct {
	rBytes  uint64
	tBytes  uint64
	rPacket uint64
	tPacket uint64
}

// StartNet 上报服务器网络流量信息
func StartNet(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()
	// 从/sys/class/net中读取全部网卡
	di, err := os.Open(sysNetDir)
	if err != nil {
		log.Panicln(err)
	}
	defer di.Close()

	ifMap := make(map[string]*netData)
	ifNames, err := di.Readdirnames(0)
	if err != nil {
		log.Panicln(err)
	}

	for _, name := range ifNames {
		res, _ := filepath.EvalSymlinks(filepath.Join(sysNetDir, name))
		// 排除虚拟网卡，如lo,docker0等
		if !strings.Contains(res, "devices/virtual") {
			ifMap[name] = new(netData)
		}
	}

	// 从/proc/net/dev中读取网卡流量
	fi, err := os.Open(netDevFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var net netInfo
	var value uint64

	firstRun := true
	ticker := time.NewTicker(time.Duration(interval.NET) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Close netInfo collector.")
			return
		case <-ticker.C:
			net.Rsum, net.Tsum = 0, 0
			net.Rpackets, net.Tpackets = 0, 0

			content := string(util.SeekAndReadAll(fi))
			pos := 0
			for i := 0; i <= len(content); i++ {
				if i != len(content) && content[i] != '\n' {
					continue
				}

				// [0]Intferace
				// [1]Rbytes [2]Rpackets  [3]Rerrs  [4]Rdrop  [5]Rfifo  [6]Rframe  [7]Rcompressed [8]Rmulticast
				// [9]Tbytes [10]Tpackets [11]Terrs [12]Tdrop [13]Tfifo [14]Tcolls [15]Tcarrier   [16]Tcompressed
				arr := strings.Fields(content[pos:i])
				pos = i + 1
				if len(arr) < 17 {
					continue
				}

				ifName := strings.TrimSuffix(arr[0], ":")
				if _, ok := ifMap[ifName]; !ok {
					continue
				}

				util.StrToNumber(arr[1], &value)
				net.Rsum += int64(value - ifMap[ifName].rBytes)
				ifMap[ifName].rBytes = value
				util.StrToNumber(arr[9], &value)
				net.Tsum += int64(value - ifMap[ifName].tBytes)
				ifMap[ifName].tBytes = value
				util.StrToNumber(arr[2], &value)
				net.Rpackets += int64(value - ifMap[ifName].rPacket)
				ifMap[ifName].rPacket = value
				util.StrToNumber(arr[10], &value)
				net.Tpackets += int64(value - ifMap[ifName].tPacket)
				ifMap[ifName].tPacket = value
			}

			net.Rrate = net.Rsum / interval.NET
			net.Trate = net.Tsum / interval.NET
			net.Rpackets = net.Rpackets / interval.NET
			net.Tpackets = net.Tpackets / interval.NET

			if firstRun {
				firstRun = false
			} else {
				select {
				case msgChan <- net:
				case <-time.After(time.Second):
					log.Println("Timeout when send netInfo to msgChan.")
				}
			}
		}
	}
}
