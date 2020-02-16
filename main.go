package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/whoisnian/aMonitor-agent/collector"
	"github.com/whoisnian/aMonitor-agent/config"
	"github.com/whoisnian/aMonitor-agent/sender"
)

// CONFIG 全局配置项
var CONFIG *config.Config

var configFilePath = flag.String("config", "config.json", "Specify a path to a custom config file")

func main() {
	// 读取agent配置
	flag.Parse()
	CONFIG = config.Load(*configFilePath)

	// 消息缓冲区
	msgChan := make(chan interface{}, 64)

	// sender监控缓冲区并发送其中数据
	sender.Init(CONFIG)
	go sender.WaitAndSend(msgChan)

	// collector开始收集数据
	collector.Init(CONFIG)
	go collector.StartBasic(msgChan)
	//go collector.StartCPU(msgChan)
	//go collector.StartRAM(msgChan)
	//go collector.StartLoad(msgChan)
	//go collector.StartNet(msgChan)

	// 额外插件信息上报接口
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			msgChan <- body
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
	http.HandleFunc("/", handler)

	go func() {
		err := http.ListenAndServe(CONFIG.ListenAddr, nil)
		if err != nil {
			log.Panicln(err)
		}
	}()

	// 拦截Interrupt信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	close(msgChan)
}
