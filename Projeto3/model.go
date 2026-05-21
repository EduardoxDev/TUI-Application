package main

import (
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	histLen  = 60
	interval = time.Second
)

type tickMsg time.Time

type metricsMsg struct {
	snap snapshot
	err  error
}

type ring struct {
	data [histLen]float64
	pos  int
	n    int
}

func (r *ring) push(v float64) {
	r.data[r.pos] = v
	r.pos = (r.pos + 1) % histLen
	if r.n < histLen {
		r.n++
	}
}

func (r *ring) slice() []float64 {
	if r.n == 0 {
		return nil
	}
	out := make([]float64, r.n)
	start := (r.pos - r.n + histLen) % histLen
	for i := range out {
		out[i] = r.data[(start+i)%histLen]
	}
	return out
}

type model struct {
	w, h    int
	col     *collector
	snap    snapshot
	cpuH    ring
	memH    ring
	sndH    ring
	rcvH    ring
	err     error
	loading bool
	sortMem bool
}

func newModel() model {
	return model{col: newCollector(), loading: true}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(doTick(), m.doCollect())
}

func doTick() tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) doCollect() tea.Cmd {
	return func() tea.Msg {
		s, err := m.col.collect()
		return metricsMsg{snap: s, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "p":
			m.sortMem = !m.sortMem
		}
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tickMsg:
		return m, tea.Batch(doTick(), m.doCollect())
	case metricsMsg:
		m.err = msg.err
		if msg.err == nil {
			m.snap = msg.snap
			m.cpuH.push(msg.snap.cpu.total)
			m.memH.push(msg.snap.mem.percent)
			if len(msg.snap.nets) > 0 {
				m.sndH.push(msg.snap.nets[0].sendRate)
				m.rcvH.push(msg.snap.nets[0].recvRate)
			}
			m.loading = false
		}
	}
	return m, nil
}

func (m model) visibleProcs() []procInfo {
	p := make([]procInfo, len(m.snap.procs))
	copy(p, m.snap.procs)
	if m.sortMem {
		sort.Slice(p, func(i, j int) bool {
			return p[i].memPct > p[j].memPct
		})
	}
	return p
}
