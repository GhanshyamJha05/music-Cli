package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GhanshyamJha05/music-Cli/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	// Look in the application's executable directory so portable yt-dlp / ffmpeg
	// placed next to the binary are discovered automatically on any OS.
	var extraPaths []string
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		extraPaths = append(extraPaths, exeDir)
	}

	// On Windows, ensure Winget package and local tool paths are added to PATH
	// so users don't get "missing dependencies" if launched from a stale shell.
	if runtime.GOOS == "windows" {
		if userPath := os.Getenv("USERPROFILE"); userPath != "" {
			wingetDir := filepath.Join(userPath, "AppData", "Local", "Microsoft", "WinGet", "Packages")
			if entries, err := os.ReadDir(wingetDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						full := filepath.Join(wingetDir, entry.Name())
						extraPaths = append(extraPaths, full)
						// Check bin subdir
						binDir := filepath.Join(full, "bin")
						if fi, err := os.Stat(binDir); err == nil && fi.IsDir() {
							extraPaths = append(extraPaths, binDir)
						}
						// Check nested extracted dirs with bin
						if subEntries, err := os.ReadDir(full); err == nil {
							for _, sub := range subEntries {
								if sub.IsDir() {
									subBin := filepath.Join(full, sub.Name(), "bin")
									if sfi, err := os.Stat(subBin); err == nil && sfi.IsDir() {
										extraPaths = append(extraPaths, subBin)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if len(extraPaths) > 0 {
		currentPath := os.Getenv("PATH")
		sep := string(os.PathListSeparator)
		os.Setenv("PATH", currentPath+sep+strings.Join(extraPaths, sep))
	}
}

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
