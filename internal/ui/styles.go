package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette — Cyan / Neon Cyan + dark
	green     = lipgloss.Color("#00E5FF") // bright electric cyan
	darkGreen = lipgloss.Color("#00838F") // deep cyan / teal border
	white     = lipgloss.Color("#FFFFFF")
	gray      = lipgloss.Color("#B3B3B3")
	darkGray  = lipgloss.Color("#282828")
	black     = lipgloss.Color("#121212")
	accent    = lipgloss.Color("#18FFFF") // vibrant accent cyan
	yellow    = lipgloss.Color("#F59B23")
	red       = lipgloss.Color("#E8445A")
	lightBlue = lipgloss.Color("#80D8FF")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(black).
				Background(green).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(white)

	dimItemStyle = lipgloss.NewStyle().
			Foreground(gray)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(green).
			Padding(1, 2)

	nowPlayingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGreen).
			Padding(0, 2)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	keyDescStyle = lipgloss.NewStyle().
			Foreground(gray)

	inputStyle = lipgloss.NewStyle().
			Foreground(white).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(green)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	badgeStyle = lipgloss.NewStyle().
			Foreground(black).
			Background(green).
			Padding(0, 1).
			Bold(true)

	playingBadge = lipgloss.NewStyle().
			Foreground(black).
			Background(accent).
			Padding(0, 1).
			Bold(true)

	pausedBadge = lipgloss.NewStyle().
			Foreground(black).
			Background(yellow).
			Padding(0, 1).
			Bold(true)
)

// Logo returns the ASCII art logo
func Logo() string {
	return lipgloss.NewStyle().
		Foreground(green).
		Bold(true).
		Render(`
 ███╗   ███╗██╗   ██╗███████╗██╗  ██████╗
 ████╗ ████║██║   ██║██╔════╝██║ ██╔════╝
 ██╔████╔██║██║   ██║███████╗██║ ██║     
 ██║╚██╔╝██║██║   ██║╚════██║██║ ██║     
 ██║ ╚═╝ ██║╚██████╔╝███████║██║ ╚██████╗
 ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═════╝`)
}

// ProgressBar renders a smooth modern progress bar
func ProgressBar(percent float64, width int) string {
	if width <= 0 {
		width = 30
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	total := float64(width) * percent
	filled := int(total)
	if filled > width {
		filled = width
	}

	var buf strings.Builder
	for i := 0; i < filled; i++ {
		buf.WriteString("━")
	}

	if filled < width {
		buf.WriteString("╸")
		for i := filled + 1; i < width; i++ {
			buf.WriteString("─")
		}
	}

	runes := []rune(buf.String())
	if len(runes) > width {
		runes = runes[:width]
	}

	headIdx := filled
	if headIdx >= len(runes) {
		headIdx = len(runes) - 1
	}

	filledStr := string(runes[:headIdx])
	var headStr, emptyStr string
	if headIdx < len(runes) {
		headStr = string(runes[headIdx : headIdx+1])
		emptyStr = string(runes[headIdx+1:])
	}

	return lipgloss.NewStyle().Foreground(accent).Bold(true).Render(filledStr) +
		lipgloss.NewStyle().Foreground(white).Bold(true).Render(headStr) +
		lipgloss.NewStyle().Foreground(darkGray).Render(emptyStr)
}

// VolumeBar renders a volume indicator bar
func VolumeBar(vol float64, width int) string {
	if width <= 0 {
		width = 20
	}
	filled := int(float64(width) * vol)
	if filled > width {
		filled = width
	}
	empty := width - filled

	var color lipgloss.TerminalColor
	switch {
	case vol >= 0.7:
		color = green
	case vol >= 0.3:
		color = yellow
	default:
		color = red
	}

	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(darkGray).Render(strings.Repeat("░", empty))
	return bar
}
