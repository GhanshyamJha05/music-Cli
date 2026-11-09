# Music-Cli

A terminal-based music player with a clean, modern interface. Manage your local music library, create playlists, search YouTube, and stream music — all from the command line.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

---

## Features

- **Local Playback** — Play MP3, WAV, and FLAC files with full controls
- **Library Management** — Add, remove, and search your music collection
- **Playlists** — Create and manage custom playlists
- **YouTube Streaming** — Search and stream from YouTube via yt-dlp
- **Genre Recommendations** — Browse curated picks across 12+ genres
- **Volume Control** — Adjustable volume with visual indicator
- **Auto-Duration** — Song length detected automatically from file
- **Rich TUI** — Beautiful terminal interface with color theme
- **Persistent Data** — Library saved to disk, survives restarts

---

## Installation

### Option 1: 1-Click Windows Setup (Easiest — All Batteries Included)

Download the single installer from [GitHub Releases](https://github.com/GhanshyamJha05/music-Cli/releases):

- **Windows Setup**: **`music-cli-windows-setup.exe`**

Double-click to install. It automatically:
- Bundles `music.exe`, `yt-dlp.exe`, and `ffmpeg.exe` (no external downloads or Go installation needed).
- Adds `music` to your system PATH so you can type `music` in PowerShell or CMD.
- Creates an optional desktop & Start Menu shortcut.

---

### Option 2: Pre-compiled Standalone Binaries

Download directly from [GitHub Releases](https://github.com/GhanshyamJha05/music-Cli/releases):

| Platform | File |
|----------|------|
| Windows (Installer) | `music-cli-windows-setup.exe` |
| Windows (Standalone) | `music-Cli-windows.exe` |
| macOS | `music-Cli-macos` |
| Linux | `music-Cli-linux` |

### Option 2: Build from Source

```bash
# Clone the repo
git clone https://github.com/GhanshyamJha05/music-Cli.git
cd music-Cli

# Build
go build -o music-Cli .

# Run
./music-Cli
```

### Option 3: Install with Go

```bash
go install github.com/GhanshyamJha05/music-Cli@latest
```

---

### Linux (Arch Linux / Ubuntu / Debian / WSL)

On Linux, `beep` / `oto` requires ALSA development headers to compile and run audio:

```bash
# Arch Linux
sudo pacman -S alsa-lib yt-dlp ffmpeg

# Ubuntu / Debian / WSL
sudo apt update && sudo apt install -y libasound2-dev yt-dlp ffmpeg
```

---

## Prerequisites

For **online streaming** (YouTube search/recommendations), you need:

```bash
# yt-dlp — YouTube download tool
# Arch Linux:
sudo pacman -S yt-dlp
# Ubuntu/Debian:
sudo apt install yt-dlp
# Or via pip:
pip install yt-dlp

# ffmpeg — Required for audio extraction
# Arch Linux:
sudo pacman -S ffmpeg
# Ubuntu/Debian:
sudo apt install ffmpeg
# macOS:
brew install ffmpeg
# Windows: winget install Gyan.FFmpeg
```

> **Note:** Local playback works without these dependencies.

---

## Usage

### Quick Start

```bash
# Run the app
./music-Cli

# Or with Go
go run .
```

### Keyboard Shortcuts

#### Global

| Key | Action |
|-----|--------|
| `Ctrl+C` | Quit |

#### Home Screen

| Key | Action |
|-----|--------|
| `1` | Library |
| `2` | Search |
| `3` | Playlists |
| `4` | Add Song |
| `5` | Online Browse |
| `6` | Recommended |
| `n` | Now Playing |
| `q` | Quit |

#### Library / Playlist Detail

| Key | Action |
|-----|--------|
| `↑`/`↓` or `k`/`j` | Navigate |
| `Enter` | Play song |
| `Space` | Play song / Pause |
| `p` | Pause/Resume |
| `s` | Stop |
| `Tab` / `>` | Skip to next track immediately |
| `Shift+Tab` / `<` | Skip to previous track |
| `K` / `Ctrl+↑` | Move selected song UP (reorder) |
| `J` / `Ctrl+↓` | Move selected song DOWN (reorder) |
| `a` | Add to playlist (Library) |
| `d` / `r` | Remove song |
| `r` | Loop mode toggle (Library) |
| `+`/`-` | Volume up/down |
| `n` | Now Playing |
| `h` | Back |

#### Now Playing

| Key | Action |
|-----|--------|
| `Space`/`p` | Pause/Resume |
| `s` | Stop |
| `>` | Next track |
| `<` | Previous track |
| `+`/`-` | Volume up/down |
| `l` | Library |
| `h` | Back |

#### Search

| Key | Action |
|-----|--------|
| `/` | Focus search input |
| `Enter` | Submit search / Play |
| `↑`/`↓` | Navigate results |
| `Esc` | Back |

#### Playlists

| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate |
| `Enter` | Open playlist |
| `c` | Create playlist |
| `d` | Delete playlist |
| `h` | Back |

#### Online Browse

| Key | Action |
|-----|--------|
| `Enter` or `/` | Focus search input |
| `↑`/`↓` | Navigate results |
| `Enter` (on result) | Stream & play |
| `s` | Save to library |
| `Esc` | Back |

#### Recommended

| Key | Action |
|-----|--------|
| `↑`/`↓` | Navigate genres |
| `Enter` | Load genre picks |
| `↑`/`↓` (in results) | Navigate tracks |
| `Enter` (on track) | Play |
| `s` / `a` | Save to library |
| `b` / `Esc` | Back to genres |
| `h` | Back to Home |

---

## Adding Songs

### Method 1: Manual Entry

1. Press `4` on the home screen
2. Fill in the fields:
   - **Title** (required)
   - **Artist**
   - **Album**
   - **Genre**
   - **File Path** (required) — Full path to audio file
   - **Duration** (optional — auto-detected if left blank)
3. Press `Enter` to submit

### Example

```
Title:    Bohemian Rhapsody
Artist:   Queen
Album:    A Night at the Opera
Genre:    Rock
FilePath: C:\Music\queen-bohemian-rhapsody.mp3
Duration: (leave blank — auto-detected)
```

### Method 2: From YouTube

1. Press `5` for Online Browse
2. Search for a song
3. Select a result
4. Press `s` to save to library

---

## Project Structure

```
music-Cli/
├── main.go                          # Entry point
├── internal/
│   ├── library/
│   │   └── library.go               # Song/Playlist model + JSON persistence
│   ├── online/
│   │   ├── ytdlp.go                 # YouTube search + download via yt-dlp
│   │   └── recommendations.go       # Genre-based recommendations
│   ├── player/
│   │   └── player.go                # Audio playback engine (beep)
│   └── ui/
│       ├── app.go                   # App model + initialization
│       ├── update.go                # Keyboard handling + state transitions
│       ├── view.go                  # Screen renderers
│       └── styles.go                # Lipgloss styles + theme
├── go.mod
├── go.sum
└── README.md
```

---

## Data Storage

Library data is stored at:

| OS | Path |
|----|------|
| Windows | `C:\Users\YOU\.music-Cli\library.json` |
| macOS/Linux | `~/.music-Cli/library.json` |

Cached downloads (auto-cleaned after 7 days):

| OS | Path |
|----|------|
| Windows | `C:\Users\YOU\.music-Cli\cache\` |
| macOS/Linux | `~/.music-Cli/cache/` |

---

## Configuration

Currently, there is no configuration file. All settings are hardcoded:

- Default volume: 70%
- Cache cleanup: 7 days
- Online search timeout: 60 seconds (retries up to 3 times with exponential backoff)
- Download timeout: 3 minutes (retries up to 3 times with exponential backoff)

---

## Development

### Requirements

- Go 1.21+
- yt-dlp (for online features)
- ffmpeg (for audio extraction)

### Build

```bash
# Local build
go build -o music-Cli .

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o music-Cli.exe .

# Build for macOS
GOOS=darwin GOARCH=arm64 go build -o music-Cli-macos .

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o music-Cli-linux .
```

### Run Tests

```bash
go test ./...
```

---

## Tech Stack

| Component | Library |
|-----------|---------|
| TUI Framework | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styling | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| Input Fields | [bubbles](https://github.com/charmbracelet/bubbles) |
| Audio Playback | [beep](https://github.com/faiface/beep) |
| YouTube | [yt-dlp](https://github.com/yt-dlp/yt-dlp) |

---

## Roadmap

- [ ] Shuffle mode
- [ ] Repeat modes (one/all/off)
- [ ] Seek (jump to position)
- [ ] Crossfade between tracks
- [ ] Album art display
- [ ] Lyrics support
- [ ] Equalizer
- [ ] Keyboard media key support
- [ ] Configuration file
- [ ] Plugin system

---

## Contributing

Contributions are welcome! Please open an issue first to discuss what you'd like to change.

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI framework
- Audio powered by [Beep](https://github.com/faiface/beep)
- Online streaming via [yt-dlp](https://github.com/yt-dlp/yt-dlp)
