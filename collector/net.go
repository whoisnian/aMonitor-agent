package collector

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/whoisnian/aMonitor-agent/util"
)

type netLoad struct {
	Rrate int64
	Trate int64
	Rsum  int64
	Tsum  int64
}

type netData struct {
	Intferace string
	Rbytes    uint64
	Tbytes    uint64
}

// StartNet 上报服务器网络流量信息
func StartNet(msgChan chan interface{}) {
	// 从/sys/class/net中读取全部网卡
	di, err := os.Open(sysNetDir)
	if err != nil {
		log.Panicln(err)
	}
	defer di.Close()

	ifMap := make(map[string]int)
	ifNames, err := di.Readdirnames(0)
	for _, name := range ifNames {
		res, _ := filepath.EvalSymlinks(filepath.Join(sysNetDir, name))
		// 排除虚拟网卡，如lo,docker0等
		if !strings.Contains(res, "devices/virtual/net") {
			ifMap[name] = 1
		}
	}

	// 从/proc/net/dev中读取网卡流量
	fi, err := os.Open(netDevFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	var netdatas []netData
	var data netData
	var load netLoad

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-ticker.C:
			load.Rsum, load.Tsum = 0, 0

			content := string(util.SeekAndReadAll(fi))
			pos := 0
			for i := 0; i <= len(content); i++ {
				if i != len(content) && content[i] != '\n' {
					continue
				}

				// [0]Intferace
				// [1]Rbytes [2]Rpackets  [3]Rerrs  [4]Rdrop  [5]Rfifo  [6]Rframe  [7]Rcompressed [8]Rmulticast
				// [9]Tbytes [10]Tpackets [11]Terrs [12]Tdrop [13]Tfifo [14]Tcolls [15]Tcarrier   [16]Tcompressed
				res := strings.Fields(content[pos:i])
				pos = i + 1
				if len(res) < 17 {
					continue
				}

				ifName := strings.TrimSuffix(res[0], ":")
				if ifMap[ifName] < 1 {
					continue
				}

				data.Intferace = ifName
				util.StrToNumber(res[1], &data.Rbytes)
				util.StrToNumber(res[9], &data.Tbytes)

				if ifMap[ifName] < 2 {
					ifMap[ifName] = len(netdatas) + 2
					netdatas = append(netdatas, data)
				} else {
					load.Rsum += int64(data.Rbytes - netdatas[ifMap[ifName]-2].Rbytes)
					load.Tsum += int64(data.Tbytes - netdatas[ifMap[ifName]-2].Tbytes)
					netdatas[ifMap[ifName]-2] = data
				}
			}

			// B ==> KB
			// load.Rsum >>= 10
			// load.Tsum >>= 10
			load.Rrate = load.Rsum / interval
			load.Trate = load.Tsum / interval

			if load.Rsum+load.Tsum > 0 {
				msgChan <- load
			}
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
