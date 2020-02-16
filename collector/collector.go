package collector

import (
	"log"

	"github.com/whoisnian/aMonitor-agent/config"
	"github.com/whoisnian/aMonitor-agent/util"
)

const statFile = "/proc/stat"
const uptimeFile = "/proc/uptime"
const loadavgFile = "/proc/loadavg"
const meminfoFile = "/proc/meminfo"
const cpuinfoFile = "/proc/cpuinfo"
const vmstatFile = "/proc/vmstat"
const osReleaseFile = "/etc/os-release"
const kernelVersionFile = "/proc/sys/kernel/osrelease"
const hostnameFile = "/proc/sys/kernel/hostname"
const netDevFile = "/proc/net/dev"

var interval int64

// Init 检查依赖文件是否存在
func Init(CONFIG *config.Config) {
	interval = CONFIG.Interval
	fileArray := []string{
		statFile,
		uptimeFile,
		loadavgFile,
		meminfoFile,
		cpuinfoFile,
		vmstatFile,
		osReleaseFile,
		kernelVersionFile,
		hostnameFile,
		netDevFile}
	for _, file := range fileArray {
		if !util.FileExists(file) {
			log.Panicf("Can't find proc file '%s'", file)
		}
	}
}
