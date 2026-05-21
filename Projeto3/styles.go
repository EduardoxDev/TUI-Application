package main

import "github.com/charmbracelet/lipgloss"

var (
	clrPurple = lipgloss.Color("#7C3AED")
	clrGreen  = lipgloss.Color("#10B981")
	clrYellow = lipgloss.Color("#F59E0B")
	clrRed    = lipgloss.Color("#EF4444")
	clrBlue   = lipgloss.Color("#3B82F6")
	clrBg     = lipgloss.Color("#111827")
	clrBorder = lipgloss.Color("#4B5563")
	clrText   = lipgloss.Color("#F9FAFB")
	clrDim    = lipgloss.Color("#6B7280")

	stPanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(clrBorder)

	stTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(clrText).
		Background(clrPurple).
		Padding(0, 1)

	stHeaderBg = lipgloss.NewStyle().
		Background(clrBg).
		Foreground(clrText)

	stSecTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(clrPurple)

	stDim = lipgloss.NewStyle().Foreground(clrDim)

	stFooter = lipgloss.NewStyle().
		Background(clrBg).
		Foreground(clrDim).
		Padding(0, 1)
)

func barColor(pct float64) lipgloss.Color {
	switch {
	case pct >= 90:
		return clrRed
	case pct >= 70:
		return clrYellow
	default:
		return clrGreen
	}
}

func memBarColor(pct float64) lipgloss.Color {
	switch {
	case pct >= 90:
		return clrRed
	case pct >= 80:
		return clrYellow
	default:
		return clrBlue
	}
}
