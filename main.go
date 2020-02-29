package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/whoisnian/aMonitor-agent/internal/collector"
	"github.com/whoisnian/aMonitor-agent/internal/config"
	"github.com/whoisnian/aMonitor-agent/internal/sender"
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

	// 使用context在程序停止时通知goroutine
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	// sender监控缓冲区并发送其中数据
	sender.Init(CONFIG)
	wg.Add(1)
	go sender.WaitAndSend(ctx, wg, msgChan)

	// collector开始收集数据
	collector.Init(CONFIG)
	collector.RunBasic(msgChan)

	wg.Add(6)
	go collector.StartCPU(ctx, wg, msgChan)
	go collector.StartMEM(ctx, wg, msgChan)
	go collector.StartLoad(ctx, wg, msgChan)
	go collector.StartNet(ctx, wg, msgChan)
	go collector.StartMounts(ctx, wg, msgChan)
	go collector.StartDisk(ctx, wg, msgChan)

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
	cancel()
	wg.Wait()
}
