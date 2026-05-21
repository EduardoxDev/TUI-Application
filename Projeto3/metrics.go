package main

import (
	"context"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type cpuStats struct {
	total float64
	cores []float64
}

type memStats struct {
	total     uint64
	used      uint64
	percent   float64
	swapTotal uint64
	swapUsed  uint64
	swapPct   float64
}

type diskStat struct {
	path    string
	total   uint64
	used    uint64
	percent float64
}

type netStat struct {
	name      string
	sendRate  float64
	recvRate  float64
	bytesSent uint64
	bytesRecv uint64
}

type procInfo struct {
	pid      int32
	name     string
	cpu      float64
	memPct   float32
	memBytes uint64
}

type snapshot struct {
	ts       time.Time
	hostname string
	platform string
	uptime   uint64
	cpu      cpuStats
	mem      memStats
	disks    []diskStat
	nets     []netStat
	procs    []procInfo
}

type collector struct {
	prevNet   map[string]gnet.IOCountersStat
	prevNetTs time.Time
	procCache map[int32]*process.Process
}

func newCollector() *collector {
	return &collector{
		prevNet:   make(map[string]gnet.IOCountersStat),
		procCache: make(map[int32]*process.Process),
	}
}

func (c *collector) collect() (snapshot, error) {
	ctx := context.Background()
	now := time.Now()
	var s snapshot
	s.ts = now

	if info, err := host.InfoWithContext(ctx); err == nil {
		s.hostname = info.Hostname
		s.platform = info.Platform
		s.uptime = info.Uptime
	}

	if pcts, err := cpu.PercentWithContext(ctx, 0, false); err == nil && len(pcts) > 0 {
		s.cpu.total = pcts[0]
	}
	if pcts, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		s.cpu.cores = pcts
	}

	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		s.mem.total = vm.Total
		s.mem.used = vm.Used
		s.mem.percent = vm.UsedPercent
	}
	if sw, err := mem.SwapMemoryWithContext(ctx); err == nil {
		s.mem.swapTotal = sw.Total
		s.mem.swapUsed = sw.Used
		s.mem.swapPct = sw.UsedPercent
	}

	if parts, err := disk.PartitionsWithContext(ctx, false); err == nil {
		for _, p := range parts {
			u, err := disk.UsageWithContext(ctx, p.Mountpoint)
			if err != nil || u.Total == 0 {
				continue
			}
			s.disks = append(s.disks, diskStat{
				path:    p.Mountpoint,
				total:   u.Total,
				used:    u.Used,
				percent: u.UsedPercent,
			})
		}
	}

	if nets, err := gnet.IOCountersWithContext(ctx, true); err == nil {
		elapsed := now.Sub(c.prevNetTs).Seconds()
		if elapsed < 0.01 {
			elapsed = 1
		}
		newPrev := make(map[string]gnet.IOCountersStat, len(nets))
		for _, n := range nets {
			newPrev[n.Name] = n
			if n.BytesSent == 0 && n.BytesRecv == 0 {
				continue
			}
			stat := netStat{
				name:      n.Name,
				bytesSent: n.BytesSent,
				bytesRecv: n.BytesRecv,
			}
			if prev, ok := c.prevNet[n.Name]; ok && !c.prevNetTs.IsZero() {
				if n.BytesSent >= prev.BytesSent {
					stat.sendRate = float64(n.BytesSent-prev.BytesSent) / elapsed
				}
				if n.BytesRecv >= prev.BytesRecv {
					stat.recvRate = float64(n.BytesRecv-prev.BytesRecv) / elapsed
				}
			}
			s.nets = append(s.nets, stat)
		}
		c.prevNet = newPrev
		c.prevNetTs = now
	}

	// Reuse process objects across collections to track CPU% deltas correctly.
	pids, _ := process.PidsWithContext(ctx)
	newCache := make(map[int32]*process.Process, len(pids))
	for _, pid := range pids {
		if p, ok := c.procCache[pid]; ok {
			newCache[pid] = p
		} else if p, err := process.NewProcessWithContext(ctx, pid); err == nil {
			newCache[pid] = p
		}
	}
	c.procCache = newCache

	for _, p := range newCache {
		cpuPct, _ := p.CPUPercentWithContext(ctx)
		name, _ := p.NameWithContext(ctx)
		memPct, _ := p.MemoryPercentWithContext(ctx)
		var memB uint64
		if mi, err := p.MemoryInfoWithContext(ctx); err == nil && mi != nil {
			memB = mi.RSS
		}
		s.procs = append(s.procs, procInfo{
			pid:      p.Pid,
			name:     name,
			cpu:      cpuPct,
			memPct:   memPct,
			memBytes: memB,
		})
	}
	sort.Slice(s.procs, func(i, j int) bool {
		return s.procs[i].cpu > s.procs[j].cpu
	})
	if len(s.procs) > 20 {
		s.procs = s.procs[:20]
	}

	return s, nil
}
