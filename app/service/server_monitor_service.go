package service

import (
	"context"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
	"gorm.io/gorm"
)

// processStartedAt 记录本进程加载服务包的时间，用来算运行时长。
var processStartedAt = time.Now()

const (
	cpuSampleInterval = 200 * time.Millisecond
	probeTimeout      = 800 * time.Millisecond
)

// ServerSnapshot 是一次运行概况。某一项失败时只填写该项的错误，其余字段仍返回。
type ServerSnapshot struct {
	CollectedAt string           `json:"collectedAt"`
	Host        HostSnapshot     `json:"host"`
	Runtime     RuntimeSnapshot  `json:"runtime"`
	Database    DatabaseSnapshot `json:"database"`
	Redis       RedisSnapshot    `json:"redis"`
}

// HostSnapshot 是主机、CPU、内存、网络和磁盘。
type HostSnapshot struct {
	Hostname         string         `json:"hostname"`
	OS               string         `json:"os"`
	Platform         string         `json:"platform"`
	PlatformVersion  string         `json:"platformVersion"`
	Arch             string         `json:"arch"`
	UptimeSeconds    uint64         `json:"uptimeSeconds"`
	HostError        string         `json:"hostError,omitempty"`
	CPUModel         string         `json:"cpuModel"`
	CPUCores         int            `json:"cpuCores"`
	CPUPercent       float64        `json:"cpuPercent"`
	CPUUser          float64        `json:"cpuUser"`
	CPUSystem        float64        `json:"cpuSystem"`
	CPUIOWait        float64        `json:"cpuIOWait"`
	CPUError         string         `json:"cpuError,omitempty"`
	MemoryTotal      uint64         `json:"memoryTotal"`
	MemoryUsed       uint64         `json:"memoryUsed"`
	MemoryAvailable  uint64         `json:"memoryAvailable"`
	MemoryPercent    float64        `json:"memoryPercent"`
	MemoryError      string         `json:"memoryError,omitempty"`
	NetRecvBytes     uint64         `json:"netRecvBytes"`
	NetSentBytes     uint64         `json:"netSentBytes"`
	NetOK            bool           `json:"netOK"`
	NetError         string         `json:"netError,omitempty"`
	DiskReadBytes    uint64         `json:"diskReadBytes"`
	DiskWriteBytes   uint64         `json:"diskWriteBytes"`
	DiskIOOK         bool           `json:"diskIOOK"`
	DiskIOError      string         `json:"diskIOError,omitempty"`
	Load1            float64        `json:"load1"`
	LoadKnown        bool           `json:"loadKnown"`
	TemperatureC     float64        `json:"temperatureC"`
	TemperatureKnown bool           `json:"temperatureKnown"`
	ProcessCount     uint64         `json:"processCount"`
	Disks            []DiskSnapshot `json:"disks"`
	DiskError        string         `json:"diskError,omitempty"`
}

// DiskSnapshot 是一个挂载点的占用。
type DiskSnapshot struct {
	Mount   string  `json:"mount"`
	FsType  string  `json:"fsType"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

// RuntimeSnapshot 是当前 Go 进程。
type RuntimeSnapshot struct {
	GoVersion     string `json:"goVersion"`
	NumCPU        int    `json:"numCPU"`
	Goroutines    int    `json:"goroutines"`
	HeapAlloc     uint64 `json:"heapAlloc"`
	HeapSys       uint64 `json:"heapSys"`
	Sys           uint64 `json:"sys"`
	NumGC         uint32 `json:"numGC"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
}

// DatabaseSnapshot 是数据库连接池和一次 Ping。
type DatabaseSnapshot struct {
	OK               bool   `json:"ok"`
	Error            string `json:"error,omitempty"`
	Open             int    `json:"open"`
	InUse            int    `json:"inUse"`
	Idle             int    `json:"idle"`
	WaitCount        int64  `json:"waitCount"`
	MaxOpen          int    `json:"maxOpen"`
	PingMilliseconds int64  `json:"pingMilliseconds"`
}

// RedisSnapshot 是 Redis INFO 里的内存、连接和命中率。
type RedisSnapshot struct {
	OK               bool    `json:"ok"`
	Error            string  `json:"error,omitempty"`
	Version          string  `json:"version"`
	UptimeSeconds    int64   `json:"uptimeSeconds"`
	UsedMemory       uint64  `json:"usedMemory"`
	UsedMemoryHuman  string  `json:"usedMemoryHuman"`
	ConnectedClients int64   `json:"connectedClients"`
	KeyspaceHits     int64   `json:"keyspaceHits"`
	KeyspaceMisses   int64   `json:"keyspaceMisses"`
	HitRate          float64 `json:"hitRate"`
	HitRateKnown     bool    `json:"hitRateKnown"`
}

// ServerMonitorService 采集主机和依赖的运行概况。
type ServerMonitorService struct {
	db    *gorm.DB
	redis redis.UniversalClient
}

// NewServerMonitorService 创建系统监控服务。
func NewServerMonitorService(db *gorm.DB, redisClient redis.UniversalClient) *ServerMonitorService {
	return &ServerMonitorService{db: db, redis: redisClient}
}

// Snapshot 采集一次。单项失败不影响其他项。
func (i *ServerMonitorService) Snapshot(ctx context.Context) ServerSnapshot {
	return ServerSnapshot{
		CollectedAt: time.Now().Format(time.RFC3339),
		Host:        readHost(ctx),
		Runtime:     readRuntime(),
		Database:    readDatabase(ctx, i.db),
		Redis:       readRedis(ctx, i.redis),
	}
}

func readRuntime() RuntimeSnapshot {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return RuntimeSnapshot{
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		Goroutines:    runtime.NumGoroutine(),
		HeapAlloc:     stats.HeapAlloc,
		HeapSys:       stats.HeapSys,
		Sys:           stats.Sys,
		NumGC:         stats.NumGC,
		UptimeSeconds: int64(time.Since(processStartedAt).Seconds()),
	}
}

func readHost(ctx context.Context) HostSnapshot {
	var snap HostSnapshot
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		snap.HostError = err.Error()
	} else {
		snap.Hostname = info.Hostname
		snap.OS = info.OS
		snap.Platform = info.Platform
		snap.PlatformVersion = info.PlatformVersion
		snap.Arch = info.KernelArch
		snap.UptimeSeconds = info.Uptime
		snap.ProcessCount = info.Procs
	}

	infos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		snap.CPUError = err.Error()
	} else if len(infos) > 0 {
		snap.CPUModel = infos[0].ModelName
	}
	cores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		snap.CPUError = joinProbeError(snap.CPUError, err)
	} else {
		snap.CPUCores = cores
	}
	readCPUTimes(ctx, &snap)

	readTraffic(ctx, &snap)
	readLoad(ctx, &snap)
	readTemperature(ctx, &snap)

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		snap.MemoryError = err.Error()
	} else {
		snap.MemoryTotal = vm.Total
		snap.MemoryUsed = vm.Used
		snap.MemoryAvailable = vm.Available
		snap.MemoryPercent = round1(vm.UsedPercent)
	}

	snap.Disks, snap.DiskError = readDisks(ctx)
	return snap
}

func readCPUTimes(ctx context.Context, snap *HostSnapshot) {
	first, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		snap.CPUError = joinProbeError(snap.CPUError, err)
		return
	}
	if err := waitSample(ctx, cpuSampleInterval); err != nil {
		snap.CPUError = joinProbeError(snap.CPUError, err)
		return
	}
	second, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		snap.CPUError = joinProbeError(snap.CPUError, err)
		return
	}
	if len(first) == 0 || len(second) == 0 {
		return
	}
	before, after := first[0], second[0]
	total := after.Total() - before.Total()
	if runtime.GOOS == "linux" {
		total -= (after.Guest - before.Guest) + (after.GuestNice - before.GuestNice)
	}
	if total <= 0 {
		return
	}
	snap.CPUUser = round1(nonNegative(after.User-before.User) / total * 100)
	snap.CPUSystem = round1(nonNegative(after.System-before.System) / total * 100)
	snap.CPUIOWait = round1(nonNegative(after.Iowait-before.Iowait) / total * 100)
	idle := nonNegative(after.Idle-before.Idle) / total * 100
	if idle > 100 {
		idle = 100
	}
	snap.CPUPercent = round1(100 - idle)
}

func waitSample(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func nonNegative(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func readTraffic(ctx context.Context, snap *HostSnapshot) {
	counters, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		snap.NetError = err.Error()
	} else if len(counters) > 0 {
		snap.NetOK = true
		snap.NetRecvBytes = counters[0].BytesRecv
		snap.NetSentBytes = counters[0].BytesSent
	}

	ioCounters, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		snap.DiskIOError = err.Error()
		return
	}
	var readBytes, writeBytes uint64
	for _, item := range ioCounters {
		readBytes += item.ReadBytes
		writeBytes += item.WriteBytes
	}
	snap.DiskIOOK = true
	snap.DiskReadBytes = readBytes
	snap.DiskWriteBytes = writeBytes
}

func readLoad(ctx context.Context, snap *HostSnapshot) {
	avg, err := load.AvgWithContext(ctx)
	if err != nil || avg == nil {
		return
	}
	snap.LoadKnown = true
	snap.Load1 = round1(avg.Load1)
}

func readTemperature(ctx context.Context, snap *HostSnapshot) {
	sensors, err := sensors.TemperaturesWithContext(ctx)
	if err != nil || len(sensors) == 0 {
		return
	}
	var sum float64
	var count int
	for _, sensor := range sensors {
		if sensor.Temperature <= 0 || sensor.Temperature >= 150 {
			continue
		}
		sum += sensor.Temperature
		count++
	}
	if count == 0 {
		return
	}
	snap.TemperatureKnown = true
	snap.TemperatureC = round1(sum / float64(count))
}

func readDisks(ctx context.Context) ([]DiskSnapshot, string) {
	parts, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return []DiskSnapshot{}, err.Error()
	}
	seen := make(map[string]struct{}, len(parts))
	disks := make([]DiskSnapshot, 0, len(parts))
	tried := 0
	for _, part := range parts {
		if skipFilesystem(part.Fstype) || part.Mountpoint == "" {
			continue
		}
		if _, ok := seen[part.Mountpoint]; ok {
			continue
		}
		seen[part.Mountpoint] = struct{}{}
		tried++
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}
		disks = append(disks, DiskSnapshot{
			Mount:   part.Mountpoint,
			FsType:  part.Fstype,
			Total:   usage.Total,
			Used:    usage.Used,
			Percent: round1(usage.UsedPercent),
		})
	}
	if tried > 0 && len(disks) == 0 {
		return []DiskSnapshot{}, "无法读取磁盘占用"
	}
	return disks, ""
}

func skipFilesystem(fsType string) bool {
	switch strings.ToLower(fsType) {
	case "autofs", "binfmt_misc", "bpf", "cgroup", "cgroup2", "configfs", "debugfs",
		"devfs", "devpts", "devtmpfs", "fusectl", "hugetlbfs", "mqueue", "nsfs",
		"overlay", "proc", "pstore", "ramfs", "securityfs", "squashfs", "sysfs",
		"tmpfs", "tracefs":
		return true
	default:
		return false
	}
}

func readDatabase(ctx context.Context, gdb *gorm.DB) DatabaseSnapshot {
	var snap DatabaseSnapshot
	if gdb == nil {
		snap.Error = "数据库未连接"
		return snap
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		snap.Error = err.Error()
		return snap
	}
	stats := sqlDB.Stats()
	snap.Open = stats.OpenConnections
	snap.InUse = stats.InUse
	snap.Idle = stats.Idle
	snap.WaitCount = stats.WaitCount
	snap.MaxOpen = stats.MaxOpenConnections

	pingCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	started := time.Now()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		snap.Error = err.Error()
		return snap
	}
	snap.PingMilliseconds = time.Since(started).Milliseconds()
	snap.OK = true
	return snap
}

func readRedis(ctx context.Context, client redis.UniversalClient) RedisSnapshot {
	var snap RedisSnapshot
	if client == nil {
		snap.Error = "Redis 未配置"
		return snap
	}
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	raw, err := client.Info(probeCtx, "server", "memory", "stats", "clients").Result()
	if err != nil {
		snap.Error = err.Error()
		return snap
	}
	info := parseRedisInfo(raw)
	snap.OK = true
	snap.Version = info["redis_version"]
	snap.UptimeSeconds = parseInt64(info["uptime_in_seconds"])
	snap.UsedMemory = parseUint64(info["used_memory"])
	snap.UsedMemoryHuman = info["used_memory_human"]
	snap.ConnectedClients = parseInt64(info["connected_clients"])
	snap.KeyspaceHits = parseInt64(info["keyspace_hits"])
	snap.KeyspaceMisses = parseInt64(info["keyspace_misses"])
	total := snap.KeyspaceHits + snap.KeyspaceMisses
	if total > 0 {
		snap.HitRateKnown = true
		snap.HitRate = round1(float64(snap.KeyspaceHits) * 100 / float64(total))
	}
	return snap
}

func parseRedisInfo(raw string) map[string]string {
	info := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		info[key] = value
	}
	return info
}

func parseInt64(value string) int64 {
	n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func parseUint64(value string) uint64 {
	n, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func joinProbeError(current string, err error) string {
	if err == nil {
		return current
	}
	if current == "" {
		return err.Error()
	}
	if strings.Contains(current, err.Error()) {
		return current
	}
	return current + "；" + err.Error()
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}
