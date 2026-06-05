// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package monitor

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// 应用启动时间
var appStartTime = time.Now()

// CPU 使用率缓存（后台 goroutine 每 3 秒采样，避免每次请求阻塞 1 秒）
var cachedCpuUsage float64

func init() {
	go func() {
		for {
			percent, err := cpu.Percent(time.Second, false)
			if err == nil && len(percent) > 0 {
				cachedCpuUsage = percent[0]
			}
			time.Sleep(3 * time.Second)
		}
	}()
}

// ServerInfo 服务器综合信息
type ServerInfo struct {
	Os      OsInfo      `json:"os"`
	Cpu     CpuInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    DiskInfo    `json:"disk"`
	Runtime RuntimeInfo `json:"runtime"`
}

// OsInfo 操作系统信息
type OsInfo struct {
	Hostname string `json:"hostname"` // 主机名
	OS       string `json:"os"`       // 操作系统
	Platform string `json:"platform"` // 平台
	Arch     string `json:"arch"`     // 架构
	GoVer    string `json:"goVer"`    // Go 版本
	Uptime   string `json:"uptime"`   // 系统运行时长
	AppTime  string `json:"appTime"`  // 应用运行时长
}

// CpuInfo CPU 信息
type CpuInfo struct {
	Cores   int     `json:"cores"`   // 核心数
	Usage   float64 `json:"usage"`   // 使用率 (%)
	ModelName string `json:"modelName"` // CPU 型号
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total       uint64  `json:"total"`       // 总量 (bytes)
	Used        uint64  `json:"used"`        // 已用 (bytes)
	Available   uint64  `json:"available"`   // 可用 (bytes)
	UsageRate   float64 `json:"usageRate"`   // 使用率 (%)
	TotalStr    string  `json:"totalStr"`    // 总量（人类可读）
	UsedStr     string  `json:"usedStr"`     // 已用（人类可读）
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total     uint64  `json:"total"`     // 总量 (bytes)
	Used      uint64  `json:"used"`      // 已用 (bytes)
	Free      uint64  `json:"free"`      // 剩余 (bytes)
	UsageRate float64 `json:"usageRate"` // 使用率 (%)
	TotalStr  string  `json:"totalStr"`
	UsedStr   string  `json:"usedStr"`
}

// RuntimeInfo Go 运行时信息
type RuntimeInfo struct {
	Goroutines int    `json:"goroutines"` // goroutine 数量
	HeapAlloc  string `json:"heapAlloc"`  // 堆内存分配
	HeapSys    string `json:"heapSys"`    // 堆内存系统分配
	StackInUse string `json:"stackInUse"` // 栈内存使用
	NumGC      uint32 `json:"numGC"`      // GC 次数
	LastGC     string `json:"lastGC"`     // 上次 GC 时间
}

// GetServerInfo 采集当前服务器信息
func GetServerInfo() (*ServerInfo, error) {
	info := &ServerInfo{}

	// OS 信息
	hostname, _ := os.Hostname()
	hostInfo, _ := host.Info()
	info.Os = OsInfo{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Platform: getPlatform(hostInfo),
		Arch:     runtime.GOARCH,
		GoVer:    runtime.Version(),
		Uptime:   formatDuration(getSystemUptime(hostInfo)),
		AppTime:  formatDuration(time.Since(appStartTime)),
	}

	// CPU 信息（使用缓存值，不再阻塞采样）
	cpuCores, _ := cpu.Counts(true)
	cpuInfos, _ := cpu.Info()
	modelName := ""
	if len(cpuInfos) > 0 {
		modelName = cpuInfos[0].ModelName
	}
	info.Cpu = CpuInfo{
		Cores:     cpuCores,
		Usage:     round2(cachedCpuUsage),
		ModelName: modelName,
	}

	// 内存信息
	memInfo, _ := mem.VirtualMemory()
	if memInfo != nil {
		info.Memory = MemoryInfo{
			Total:     memInfo.Total,
			Used:      memInfo.Used,
			Available: memInfo.Available,
			UsageRate: round2(memInfo.UsedPercent),
			TotalStr:  formatBytes(memInfo.Total),
			UsedStr:   formatBytes(memInfo.Used),
		}
	}

	// 磁盘信息（根分区或 C 盘）
	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:"
	}
	diskInfo, _ := disk.Usage(diskPath)
	if diskInfo != nil {
		info.Disk = DiskInfo{
			Total:     diskInfo.Total,
			Used:      diskInfo.Used,
			Free:      diskInfo.Free,
			UsageRate: round2(diskInfo.UsedPercent),
			TotalStr:  formatBytes(diskInfo.Total),
			UsedStr:   formatBytes(diskInfo.Used),
		}
	}

	// Go Runtime 信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	lastGC := "N/A"
	if m.LastGC > 0 {
		lastGC = time.Unix(0, int64(m.LastGC)).Format("2006-01-02 15:04:05")
	}
	info.Runtime = RuntimeInfo{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  formatBytes(m.HeapAlloc),
		HeapSys:    formatBytes(m.HeapSys),
		StackInUse: formatBytes(m.StackInuse),
		NumGC:      m.NumGC,
		LastGC:     lastGC,
	}

	return info, nil
}

// --------- 辅助函数 ---------

func getPlatform(info *host.InfoStat) string {
	if info != nil {
		return fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion)
	}
	return runtime.GOOS
}

func getSystemUptime(info *host.InfoStat) time.Duration {
	if info != nil {
		return time.Duration(info.Uptime) * time.Second
	}
	return 0
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	fb := float64(b)
	div, exp := float64(unit), 0
	for n := fb / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", fb/div, units[exp])
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
