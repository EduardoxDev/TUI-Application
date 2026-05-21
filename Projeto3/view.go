package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.w < 60 || m.h < 10 {
		return fmt.Sprintf("\n  Terminal very pequeno (%dx%d). Mínimo: 60x10.\n", m.w, m.h)
	}
	if m.loading {
		return "\n\n  Coletando métricas...\n"
	}
	if m.err != nil {
		return fmt.Sprintf("\n  Erro: %v\n  Pressione q para sair.\n", m.err)
	}

	lw := m.w / 2
	rw := m.w - lw

	// RoundedBorder adds 1 char per side = 2 total; inner = panelWidth - 2
	lIn := lw - 2
	rIn := rw - 2
	fIn := m.w - 2

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		lipgloss.JoinHorizontal(lipgloss.Top,
			stPanel.Width(lIn).Render(m.renderCPU(lIn)),
			stPanel.Width(rIn).Render(m.renderMem(rIn)),
		),
		lipgloss.JoinHorizontal(lipgloss.Top,
			stPanel.Width(lIn).Render(m.renderDisk(lIn)),
			stPanel.Width(rIn).Render(m.renderNet(rIn)),
		),
		stPanel.Width(fIn).Render(m.renderProcs(fIn)),
		m.renderFooter(),
	)
}

// ─── Header ──────────────────────────────────────────────────────────────────

func (m model) renderHeader() string {
	title := stTitle.Render(" SYSMON ")
	clock := stHeaderBg.Render(time.Now().Format(" 15:04:05 "))
	info := fmt.Sprintf("  %s  ·  %s  ·  up %s",
		m.snap.hostname, m.snap.platform, fmtUptime(m.snap.uptime))

	titleW := lipgloss.Width(title)
	clockW := lipgloss.Width(clock)
	midW := m.w - titleW - clockW
	if midW < 0 {
		midW = 0
		info = ""
	}

	mid := stHeaderBg.Width(midW).MaxWidth(midW).Render(info)
	return title + mid + clock
}

// ─── Footer ──────────────────────────────────────────────────────────────────

func (m model) renderFooter() string {
	sortLabel := "CPU"
	if m.sortMem {
		sortLabel = "Mem"
	}
	keys := fmt.Sprintf("[q] Sair   [p] Alternar ordenação (atual: %s)", sortLabel)
	return stFooter.Width(m.w).Render(keys)
}

// ─── CPU ─────────────────────────────────────────────────────────────────────

func (m model) renderCPU(w int) string {
	lines := []string{stSecTitle.Render("  CPU")}

	pct := m.snap.cpu.total
	barW := w - 16
	if barW < 2 {
		barW = 2
	}
	bar := coloredBar(pct, barW, barColor(pct))
	lines = append(lines, fmt.Sprintf("  Total  %s %5.1f%%", bar, pct))

	if hist := m.cpuH.slice(); len(hist) > 0 {
		spk := sparkline(hist, barW+8, 100.0, barColor(pct))
		lines = append(lines, "  "+spk)
	}
	lines = append(lines, "")

	coreBarW := w - 18
	if coreBarW < 2 {
		coreBarW = 2
	}
	for i, c := range m.snap.cpu.cores {
		if i >= 8 {
			lines = append(lines, stDim.Render(fmt.Sprintf("  ... +%d cores", len(m.snap.cpu.cores)-8)))
			break
		}
		b := coloredBar(c, coreBarW, barColor(c))
		lines = append(lines, fmt.Sprintf("  Core%-2d %s %5.1f%%", i, b, c))
	}

	return strings.Join(lines, "\n")
}

// ─── Memory ──────────────────────────────────────────────────────────────────

func (m model) renderMem(w int) string {
	lines := []string{stSecTitle.Render("  MEMÓRIA")}

	mn := m.snap.mem
	barW := w - 16
	if barW < 2 {
		barW = 2
	}

	ramBar := coloredBar(mn.percent, barW, memBarColor(mn.percent))
	lines = append(lines, fmt.Sprintf("  RAM    %s %5.1f%%", ramBar, mn.percent))
	lines = append(lines, stDim.Render(fmt.Sprintf("         %s / %s",
		fmtBytes(mn.used), fmtBytes(mn.total))))

	if hist := m.memH.slice(); len(hist) > 0 {
		spk := sparkline(hist, barW+8, 100.0, memBarColor(mn.percent))
		lines = append(lines, "  "+spk)
	}

	if mn.swapTotal > 0 {
		lines = append(lines, "")
		swapBar := coloredBar(mn.swapPct, barW, memBarColor(mn.swapPct))
		lines = append(lines, fmt.Sprintf("  Swap   %s %5.1f%%", swapBar, mn.swapPct))
		lines = append(lines, stDim.Render(fmt.Sprintf("         %s / %s",
			fmtBytes(mn.swapUsed), fmtBytes(mn.swapTotal))))
	}

	return strings.Join(lines, "\n")
}

// ─── Disk ────────────────────────────────────────────────────────────────────

func (m model) renderDisk(w int) string {
	lines := []string{stSecTitle.Render("  DISCO")}

	if len(m.snap.disks) == 0 {
		lines = append(lines, stDim.Render("  Nenhum disco encontrado"))
		return strings.Join(lines, "\n")
	}

	barW := w - 15
	if barW < 2 {
		barW = 2
	}
	for i, d := range m.snap.disks {
		if i >= 6 {
			break
		}
		path := d.path
		if len(path) > 8 {
			path = path[:5] + "..."
		}
		bar := coloredBar(d.percent, barW, barColor(d.percent))
		lines = append(lines, fmt.Sprintf("  %-8s %s", path, bar))
		lines = append(lines, stDim.Render(fmt.Sprintf("           %s / %s  %.0f%%",
			fmtBytes(d.used), fmtBytes(d.total), d.percent)))
	}

	return strings.Join(lines, "\n")
}

// ─── Network ─────────────────────────────────────────────────────────────────

func (m model) renderNet(w int) string {
	lines := []string{stSecTitle.Render("  REDE")}

	if len(m.snap.nets) == 0 {
		lines = append(lines, stDim.Render("  Nenhuma interface ativa"))
		return strings.Join(lines, "\n")
	}

	shown := m.snap.nets
	if len(shown) > 3 {
		shown = shown[:3]
	}

	for _, n := range shown {
		name := n.name
		if len(name) > w-4 {
			name = name[:w-7] + "..."
		}
		lines = append(lines,
			lipgloss.NewStyle().Bold(true).Foreground(clrText).Render("  "+name))
		upColor := lipgloss.NewStyle().Foreground(clrGreen)
		downColor := lipgloss.NewStyle().Foreground(clrBlue)
		lines = append(lines, fmt.Sprintf("  %s %-14s  %s %s",
			upColor.Render("↑"),
			fmtRate(n.sendRate),
			downColor.Render("↓"),
			fmtRate(n.recvRate)))
	}

	// Sparklines for the first interface
	sndH := m.sndH.slice()
	rcvH := m.rcvH.slice()
	if len(sndH) > 0 && len(rcvH) > 0 {
		maxRate := maxSlice(append(append([]float64{}, sndH...), rcvH...))
		if maxRate == 0 {
			maxRate = 1
		}
		lineW := (w - 6) / 2
		if lineW < 4 {
			lineW = 4
		}
		lines = append(lines, "")
		sSpk := sparkline(sndH, lineW, maxRate, clrGreen)
		rSpk := sparkline(rcvH, lineW, maxRate, clrBlue)
		lines = append(lines,
			stDim.Render("↑ ")+sSpk+stDim.Render("  ↓ ")+rSpk)
	}

	return strings.Join(lines, "\n")
}

// ─── Processes ───────────────────────────────────────────────────────────────

func (m model) renderProcs(w int) string {
	sortLabel := "CPU"
	if m.sortMem {
		sortLabel = "Mem"
	}
	title := stSecTitle.Render(fmt.Sprintf("  PROCESSOS  [ordenado por %s]", sortLabel))

	nameW := w - 38
	if nameW < 8 {
		nameW = 8
	}
	if nameW > 35 {
		nameW = 35
	}

	pidS := fmt.Sprintf("%6s", "PID")
	nameS := fmt.Sprintf("%-*s", nameW, "NOME")
	thead := stDim.Render(fmt.Sprintf("  %s  %s  %6s  %5s  %10s",
		pidS, nameS, "CPU%", "MEM%", "MEM"))

	rows := []string{title, thead}

	for i, p := range m.visibleProcs() {
		if i >= 10 {
			break
		}
		name := p.name
		if len(name) > nameW {
			name = name[:nameW-1] + "…"
		}

		cpuStr := fmt.Sprintf("%6.1f", p.cpu)
		memStr := fmt.Sprintf("%5.1f", p.memPct)
		memBStr := fmt.Sprintf("%10s", fmtBytes(p.memBytes))

		cpuStyled := lipgloss.NewStyle().Foreground(barColor(p.cpu)).Render(cpuStr)
		memStyled := lipgloss.NewStyle().Foreground(memBarColor(float64(p.memPct))).Render(memStr)

		row := fmt.Sprintf("  %6d  %-*s  %s  %s  %s",
			p.pid, nameW, name, cpuStyled, memStyled, stDim.Render(memBStr))
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func coloredBar(pct float64, width int, color lipgloss.Color) string {
	if width < 1 {
		return ""
	}
	filled := int(math.Round(pct / 100.0 * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	filledS := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled))
	emptyS := stDim.Render(strings.Repeat("░", width-filled))
	return filledS + emptyS
}

func sparkline(data []float64, width int, maxVal float64, color lipgloss.Color) string {
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	if len(data) > width {
		data = data[len(data)-width:]
	}
	var sb strings.Builder
	for _, v := range data {
		idx := 0
		if maxVal > 0 {
			idx = int(v / maxVal * float64(len(blocks)-1))
		}
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		sb.WriteRune(blocks[idx])
	}
	result := lipgloss.NewStyle().Foreground(color).Render(sb.String())
	if pad := width - len(data); pad > 0 {
		result = strings.Repeat(" ", pad) + result
	}
	return result
}

func maxSlice(s []float64) float64 {
	var m float64
	for _, v := range s {
		if v > m {
			m = v
		}
	}
	return m
}

func fmtBytes(b uint64) string {
	const unit = uint64(1024)
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func fmtRate(bps float64) string {
	switch {
	case bps < 1024:
		return fmt.Sprintf("%.0f B/s", bps)
	case bps < 1024*1024:
		return fmt.Sprintf("%.1f KB/s", bps/1024)
	default:
		return fmt.Sprintf("%.2f MB/s", bps/1024/1024)
	}
}

func fmtUptime(secs uint64) string {
	d := secs / 86400
	h := (secs % 86400) / 3600
	mn := (secs % 3600) / 60
	if d > 0 {
		return fmt.Sprintf("%dd %dh %dm", d, h, mn)
	}
	return fmt.Sprintf("%dh %dm", h, mn)
}
