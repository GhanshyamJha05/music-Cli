package ui

import (
	"context"
	"sync/atomic"
	"time"

	"musiccli/internal/library"
	"musiccli/internal/online"
	"musiccli/internal/player"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen identifiers
type Screen int

const (
	ScreenNone Screen = iota - 1
	ScreenHome
	ScreenLibrary
	ScreenSearch
	ScreenPlaylists
	ScreenPlaylistDetail
	ScreenNowPlaying
	ScreenAddSong
	ScreenCreatePlaylist
	ScreenAddToPlaylist
	ScreenOnlineSearch  // online browse entry
	ScreenOnlineResults // online results list
	ScreenRecommended   // genre picker
)

// tickMsg is sent every second to update the progress bar
type tickMsg time.Time

// App is the root bubbletea model
type App struct {
	screen     Screen
	prevScreen Screen
	lib        *library.Library
	pl         *player.Player

	// ── Per-screen cursors (never share a cursor across screens) ───────────
	libCursor         int
	searchCursor      int
	playlistCursor    int
	playlistDetCursor int
	onlineCursor      int
	recCursor         int // genre picker cursor
	addToPLCursor     int

	// currently viewed playlist
	activePlaylist *library.Playlist

	// currently playing
	nowPlayingSong *library.Song
	queueSongs     []library.Song
	queueIndex     int

	// local search
	searchInput   textinput.Model
	searchResults []library.Song
	searchActive  bool

	// add song form
	addSongFields []textinput.Model
	addSongFocus  int

	// create playlist
	createPLInput textinput.Model

	// add to playlist selection
	addToPLSongID string

	// status messages
	statusMsg     string
	statusIsError bool
	statusTimer   int

	// terminal size
	width  int
	height int

	// ── Online search ──────────────────────────────────────────────────────
	onlineInput   textinput.Model
	onlineResults []online.OnlineTrack
	onlineLoading bool
	onlineErr     string

	// ── Recommended ────────────────────────────────────────────────────────
	recActiveGenre *online.Genre
	recResults     []online.OnlineTrack // separate from onlineResults
	recLoading     bool
	recErr         string
	recResCursor   int // cursor inside rec results list

	// ── Program reference for event-driven messages ────────────────────────
	program atomic.Pointer[tea.Program]
}

// NewApp initializes the TUI application
func NewApp() *App {
	lib, loadErr := library.Load()
	if loadErr != nil {
		lib = &library.Library{}
	}
	pl := player.Get()

	// Search input
	si := textinput.New()
	si.Placeholder = "Search songs, artists, albums..."
	si.CharLimit = 100

	// Add song form fields: Title, Artist, Album, Genre, FilePath, Duration
	fields := make([]textinput.Model, 6)
	placeholders := []string{"Title", "Artist", "Album", "Genre", "File Path (.mp3/.wav/.flac)", "Duration (leave blank to auto-detect)"}
	for i := range fields {
		fields[i] = textinput.New()
		fields[i].Placeholder = placeholders[i]
		fields[i].CharLimit = 200
	}
	fields[0].Focus()

	// Playlist name input
	pli := textinput.New()
	pli.Placeholder = "Playlist name..."
	pli.CharLimit = 80

	// Online search input
	oi := textinput.New()
	oi.Placeholder = "Search YouTube..."
	oi.CharLimit = 120

	// Clean old cache files on startup (10s timeout, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		online.CleanCache(ctx)
	}()

	app := &App{
		screen:        ScreenHome,
		prevScreen:    ScreenNone,
		lib:           lib,
		pl:            pl,
		searchInput:   si,
		addSongFields: fields,
		createPLInput: pli,
		onlineInput:   oi,
	}

	// When current track finishes, auto-advance queue via event-driven message
	pl.OnFinish = func() {
		if p := app.program.Load(); p != nil {
			p.Send(songFinishedMsg{})
		}
	}

	if loadErr != nil {
		app.setStatus("Warning: could not load library: "+loadErr.Error(), true)
	}

	return app
}

// SetProgram stores the tea.Program reference so OnFinish can send messages
func (a *App) SetProgram(p *tea.Program) {
	a.program.Store(p)
}

// Init starts the tick loop only if something needs periodic updates
func (a *App) Init() tea.Cmd {
	if a.needsTick() {
		return tickCmd()
	}
	return nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// setStatus sets a temporary status message and returns a tick cmd to count it down
func (a *App) setStatus(msg string, isError bool) tea.Cmd {
	a.statusMsg = msg
	a.statusIsError = isError
	a.statusTimer = 4
	return tickCmd()
}

// needsTick returns true if the app needs periodic tick updates
func (a *App) needsTick() bool {
	if a.statusTimer > 0 {
		return true
	}
	if a.nowPlayingSong != nil && a.pl.IsPlaying() {
		return true
	}
	return false
}


