package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette — Spotify-ish green + dark
	green     = lipgloss.Color("#1DB954")
	darkGreen = lipgloss.Color("#158a3e")
	white     = lipgloss.Color("#FFFFFF")
	gray      = lipgloss.Color("#B3B3B3")
	darkGray  = lipgloss.Color("#282828")
	black     = lipgloss.Color("#121212")
	accent    = lipgloss.Color("#1ed760")
	yellow    = lipgloss.Color("#F59B23")
	red       = lipgloss.Color("#E8445A")
	lightBlue = lipgloss.Color("#5FB8FF")

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
				Bold(true).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(white).
			Padding(0, 1)

	dimItemStyle = lipgloss.NewStyle().
			Foreground(gray).
			Padding(0, 1)

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
 ██████╗██╗     ██╗    ███╗   ███╗██╗   ██╗███████╗██╗ ██████╗
██╔════╝██║     ██║    ████╗ ████║██║   ██║██╔════╝██║██╔════╝
██║     ██║     ██║    ██╔████╔██║██║   ██║███████╗██║██║     
██║     ██║     ██║    ██║╚██╔╝██║██║   ██║╚════██║██║██║     
╚██████╗███████╗██║    ██║ ╚═╝ ██║╚██████╔╝███████║██║╚██████╗
 ╚═════╝╚══════╝╚═╝    ╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝ ╚═════╝`)
}

// ProgressBar renders a simple text progress bar
func ProgressBar(percent float64, width int) string {
	if width <= 0 {
		width = 30
	}
	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}
	empty := width - filled
	bar := lipgloss.NewStyle().Foreground(green).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(darkGray).Render(strings.Repeat("░", empty))
	return bar
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
