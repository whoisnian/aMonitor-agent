package main

import (
	"context"
	"flag"
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

	wg.Add(7)
	go collector.StartCPU(ctx, wg, msgChan)
	go collector.StartMEM(ctx, wg, msgChan)
	go collector.StartLoad(ctx, wg, msgChan)
	go collector.StartNet(ctx, wg, msgChan)
	go collector.StartMounts(ctx, wg, msgChan)
	go collector.StartDisk(ctx, wg, msgChan)
	go collector.StartPluginListen(ctx, wg, msgChan)

	// 拦截Interrupt信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	// 通知goroutine结束
	cancel()

	// 等待关键goroutine结束
	wg.Wait()
}
