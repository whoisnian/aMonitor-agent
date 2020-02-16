package collector

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/whoisnian/aMonitor-agent/util"
)

type basicInfo struct {
	Distro   string // 发行版名称
	Kernel   string // 内核版本
	Hostname string // 主机名
	CPUModel string // CPU型号
	CPUCores int64  // CPU核心数
}

// StartBasic 上报服务器基本信息
func StartBasic(msgChan chan interface{}) {
	var basic basicInfo

	// 从/etc/os-release中读取发行版名称
	fiOS, err := os.Open(osReleaseFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fiOS.Close()
	content := string(util.ReadAll(fiOS))

	pos := 0
	for i := 0; i <= len(content); i++ {
		if i != len(content) && content[i] != '\n' {
			continue
		}

		res := strings.SplitN(content[pos:i], "=", 2)
		pos = i + 1

		switch res[0] {
		case "PRETTY_NAME":
			basic.Distro = strings.Trim(res[1], "\"")
			break
		}
	}

	// 从/proc/sys/kernel/osrelease中读取内核版本
	fiKV, err := os.Open(kernelVersionFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fiKV.Close()

	content = string(util.ReadAll(fiKV))
	basic.Kernel = strings.TrimSpace(content)

	// 从/proc/sys/kernel/hostname中读取主机名
	fiHN, err := os.Open(hostnameFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fiHN.Close()

	content = string(util.ReadAll(fiHN))
	basic.Hostname = strings.TrimSpace(content)

	// 从/proc/cpuinfo中读取CPU型号
	fiCPU, err := os.Open(cpuinfoFile)
	if err != nil {
		log.Panicln(err)
	}
	defer fiCPU.Close()
	content = string(util.ReadAll(fiCPU))

	pos = 0
	for i := 0; i <= len(content); i++ {
		if i != len(content) && content[i] != '\n' {
			continue
		}

		res := strings.SplitN(content[pos:i], ":", 2)
		pos = i + 1

		switch strings.TrimSpace(res[0]) {
		case "model name":
			basic.CPUModel = strings.TrimSpace(res[1])
		case "cpu cores":
			basic.CPUCores, _ = strconv.ParseInt(strings.TrimSpace(res[1]), 10, 64)
		}
	}

	msgChan <- basic
}
