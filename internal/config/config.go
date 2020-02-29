package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/whoisnian/aMonitor-agent/internal/util"
)

// Interval 各项监控刷新间隔，单位为秒
type Interval struct {
	CPU    int64
	MEM    int64
	LOAD   int64
	NET    int64
	MOUNTS int64
	DISK   int64
}

// Config agent配置项
type Config struct {
	StroageURL string   // aMonitor-stroage服务器websocket上报地址，如"ws://127.0.0.1:3000"
	SourceURL  []string // aMonitor-source服务器资源下载地址，如"http://192.168.1.1:8080"
	ListenAddr string   // 本地监听地址，为插件提供上报消息的接口，如"127.0.0.1:8008"
	Interval   Interval // 各项监控刷新间隔，单位为秒
	Token      string   // 身份令牌，agent首次启动时自动获取，用来区分不同agent
}

// Load 加载配置文件
func Load(path string) *Config {
	fi, err := os.Open(path)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	content := util.ReadAll(fi)
	config := new(Config)
	err = json.Unmarshal(content, config)
	if err != nil {
		log.Panicln(err)
	}
	return config
}

// Save 保存配置文件
func Save(path string, config *Config) {
	fi, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		log.Panicln(err)
	}
	defer fi.Close()

	content, err := json.MarshalIndent(*config, "", "  ")
	if err != nil {
		log.Panicln(err)
	}

	_, err = fi.Write(content)
	if err != nil {
		log.Panicln(err)
	}
}
