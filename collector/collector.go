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
const mountsFile = "/proc/mounts"
const diskstatsFile = "/proc/diskstats"
const sysNetDir = "/sys/class/net/"

var interval config.Interval

// Init 检查依赖是否存在
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
		netDevFile,
		mountsFile,
		diskstatsFile}
	for _, file := range fileArray {
		if !util.FileExists(file) {
			log.Panicf("Can't find file '%s'", file)
		}
	}

	dirArray := []string{sysNetDir}
	for _, dir := range dirArray {
		if !util.DirExists(dir) {
			log.Panicf("Can't find dir '%s'", dir)
		}
	}
}
