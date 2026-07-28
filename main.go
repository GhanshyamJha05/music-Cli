package main

import (
	"fmt"
	"os"

	"musiccli/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	app := ui.NewApp()

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
