package ui

import (
	"github.com/GhanshyamJha05/music-Cli/internal/library"
	"github.com/GhanshyamJha05/music-Cli/internal/online"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ── Async messages ────────────────────────────────────────────────────────────

// onlineResultsMsg carries results back from a goroutine search
type onlineResultsMsg struct {
	tracks []online.OnlineTrack
	err    error
}

// onlinePlayReadyMsg signals a yt-dlp download is done
type onlinePlayReadyMsg struct {
	tempPath string
	track    online.OnlineTrack
	err      error
}

// recResultsMsg carries recommended genre results separately
type recResultsMsg struct {
	tracks []online.OnlineTrack
	err    error
}

// songFinishedMsg is sent by the player's OnFinish callback
type songFinishedMsg struct{}

// savedTrackReadyMsg signals a saved online track has been re-downloaded
type savedTrackReadyMsg struct {
	tempPath string
	song     library.Song
	err      error
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		if a.statusTimer > 0 {
			a.statusTimer--
			if a.statusTimer == 0 {
				a.statusMsg = ""
			}
		}
		if a.needsTick() {
			return a, tickCmd()
		}
		return a, nil

	case songFinishedMsg:
		if cmd := a.advanceQueue(); cmd != nil {
			return a, cmd
		}
		return a, nil

	case savedTrackReadyMsg:
		a.onlineLoading = false
		a.recLoading = false
		if msg.err != nil {
			return a, a.setStatus("Download error: "+msg.err.Error(), true)
		}
		if err := a.pl.PlayTempFile(msg.tempPath); err != nil {
			return a, a.setStatus("Playback error: "+err.Error(), true)
		}
		s := msg.song
		a.nowPlayingSong = &s
		a.queueSongs = []library.Song{s}
		a.queueIndex = 0
		a.screen = ScreenNowPlaying
		return a, tickCmd()

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case onlineResultsMsg:
		a.onlineLoading = false
		if msg.err != nil {
			a.onlineErr = msg.err.Error()
		} else {
			a.onlineErr = ""
			a.onlineResults = msg.tracks
			a.onlineCursor = 0
		}
		return a, nil

	case recResultsMsg:
		a.recLoading = false
		if msg.err != nil {
			a.recErr = msg.err.Error()
		} else {
			a.recErr = ""
			a.recResults = msg.tracks
			a.recResCursor = 0
		}
		return a, nil

	case onlinePlayReadyMsg:
		a.onlineLoading = false
		a.recLoading = false
		if msg.err != nil {
			return a, a.setStatus("Play error: "+msg.err.Error(), true)
		}
		var playErr error
		playErr = a.pl.PlayTempFile(msg.tempPath)
		if playErr != nil {
			return a, a.setStatus("Play error: "+playErr.Error(), true)
		}
		s := online.TrackToSong(msg.track)
		a.nowPlayingSong = &s
		a.queueSongs = []library.Song{s}
		a.queueIndex = 0
		a.screen = ScreenNowPlaying
		return a, tickCmd()

	case tea.KeyMsg:
		return a.handleKey(msg)
	}
	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		a.pl.Stop()
		return a, tea.Quit
	}

	// Delegate to screen-specific handlers
	switch a.screen {
	case ScreenHome:
		return a.updateHome(key)
	case ScreenLibrary:
		return a.updateLibrary(key)
	case ScreenSearch:
		return a.updateSearch(msg)
	case ScreenPlaylists:
		return a.updatePlaylists(key)
	case ScreenPlaylistDetail:
		return a.updatePlaylistDetail(key)
	case ScreenNowPlaying:
		return a.updateNowPlaying(key)
	case ScreenAddSong:
		return a.updateAddSong(msg)
	case ScreenCreatePlaylist:
		return a.updateCreatePlaylist(msg)
	case ScreenAddToPlaylist:
		return a.updateAddToPlaylist(key)
	case ScreenOnlineSearch:
		return a.updateOnlineSearch(msg)
	case ScreenOnlineResults:
		return a.updateOnlineResults(key)
	case ScreenRecommended:
		return a.updateRecommended(key)
	}
	return a, nil
}

// ── Home ────────────────────────────────────────────────────────────────────

func (a *App) updateHome(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		a.screen = ScreenLibrary
	case "2":
		a.screen = ScreenSearch
		a.searchInput.Focus()
		a.searchResults = nil
		a.searchCursor = 0
	case "3":
		a.screen = ScreenPlaylists
		a.playlistCursor = 0
	case "4":
		a.screen = ScreenAddSong
		a.resetAddSongForm()
	case "5":
		// only enter if not already loading
		if !a.onlineLoading {
			a.screen = ScreenOnlineSearch
			a.onlineResults = nil
			a.onlineErr = ""
			a.onlineInput.SetValue("")
			a.onlineInput.Focus()
			a.onlineCursor = 0
		}
	case "6":
		a.screen = ScreenRecommended
		a.recCursor = 0
		a.recActiveGenre = nil
		a.recResults = nil
		a.recErr = ""
		a.recLoading = false
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenHome
			a.screen = ScreenNowPlaying
		}
	case "q":
		a.pl.Stop()
		return a, tea.Quit
	}
	return a, nil
}

// ── Library ──────────────────────────────────────────────────────────────────

func (a *App) updateLibrary(key string) (tea.Model, tea.Cmd) {
	songs := a.lib.Songs
	switch key {
	case "up", "k":
		if a.libCursor > 0 {
			a.libCursor--
		}
	case "down", "j":
		if a.libCursor < len(songs)-1 {
			a.libCursor++
		}
	case "enter", " ":
		if len(songs) > 0 {
			a.prevScreen = ScreenLibrary
			a.screen = ScreenNowPlaying
			return a, a.playFromList(songs, a.libCursor)
		}
	case "p":
		a.pl.Pause()
	case "s":
		a.pl.Stop()
		a.nowPlayingSong = nil
	case "plus", "=":
		a.pl.VolumeUp()
	case "minus":
		a.pl.VolumeDown()
	case "d":
		if len(songs) > 0 {
			if _, err := a.lib.RemoveSong(songs[a.libCursor].ID); err != nil {
				return a, a.setStatus("Save error: "+err.Error(), true)
			}
			if a.libCursor >= len(a.lib.Songs) && a.libCursor > 0 {
				a.libCursor--
			}
			return a, a.setStatus("Song removed", false)
		}
	case "a":
		if len(songs) > 0 {
			a.addToPLSongID = songs[a.libCursor].ID
			a.prevScreen = ScreenLibrary
			a.addToPLCursor = 0
			a.screen = ScreenAddToPlaylist
		}
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenLibrary
			a.screen = ScreenNowPlaying
		}
	case "esc", "h":
		a.screen = ScreenHome
	}
	return a, nil
}

// ── Search ───────────────────────────────────────────────────────────────────

func (a *App) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if a.searchActive {
		switch key {
		case "esc":
			a.searchActive = false
			a.searchInput.Blur()
		case "enter":
			q := strings.TrimSpace(a.searchInput.Value())
			if q != "" {
				a.searchResults = a.lib.Search(q)
				a.searchCursor = 0
			}
			a.searchActive = false
			a.searchInput.Blur()
		default:
			var cmd tea.Cmd
			a.searchInput, cmd = a.searchInput.Update(msg)
			return a, cmd
		}
		return a, nil
	}

	// browsing results
	switch key {
	case "/":
		a.searchActive = true
		a.searchInput.Focus()
	case "up", "k":
		if a.searchCursor > 0 {
			a.searchCursor--
		}
	case "down", "j":
		if len(a.searchResults) > 0 && a.searchCursor < len(a.searchResults)-1 {
			a.searchCursor++
		}
	case "enter":
		if len(a.searchResults) > 0 && a.searchCursor < len(a.searchResults) {
			a.prevScreen = ScreenSearch
			a.screen = ScreenNowPlaying
			return a, a.playFromList(a.searchResults, a.searchCursor)
		}
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenSearch
			a.screen = ScreenNowPlaying
		}
	case "esc", "h":
		a.screen = ScreenHome
	}
	return a, nil
}

// ── Playlists ────────────────────────────────────────────────────────────────

func (a *App) updatePlaylists(key string) (tea.Model, tea.Cmd) {
	pls := a.lib.Playlists
	switch key {
	case "up", "k":
		if a.playlistCursor > 0 {
			a.playlistCursor--
		}
	case "down", "j":
		if a.playlistCursor < len(pls)-1 {
			a.playlistCursor++
		}
	case "enter":
		if len(pls) > 0 && a.playlistCursor < len(pls) {
			a.activePlaylist = &a.lib.Playlists[a.playlistCursor]
			a.playlistDetCursor = 0
			a.screen = ScreenPlaylistDetail
		}
	case "c":
		a.screen = ScreenCreatePlaylist
		a.createPLInput.SetValue("")
		a.createPLInput.Focus()
	case "d":
		if len(pls) > 0 && a.playlistCursor < len(pls) {
			if _, err := a.lib.DeletePlaylist(pls[a.playlistCursor].ID); err != nil {
				return a, a.setStatus("Save error: "+err.Error(), true)
			}
			if a.playlistCursor >= len(a.lib.Playlists) && a.playlistCursor > 0 {
				a.playlistCursor--
			}
			return a, a.setStatus("Playlist deleted", false)
		}
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenPlaylists
			a.screen = ScreenNowPlaying
		}
	case "esc", "h":
		a.screen = ScreenHome
	}
	return a, nil
}

// ── Playlist Detail ───────────────────────────────────────────────────────────

func (a *App) updatePlaylistDetail(key string) (tea.Model, tea.Cmd) {
	if a.activePlaylist == nil {
		a.screen = ScreenPlaylists
		return a, nil
	}
	songs := a.lib.GetPlaylistSongs(a.activePlaylist.ID)
	switch key {
	case "up", "k":
		if a.playlistDetCursor > 0 {
			a.playlistDetCursor--
		}
	case "down", "j":
		if a.playlistDetCursor < len(songs)-1 {
			a.playlistDetCursor++
		}
	case "enter", " ":
		if len(songs) > 0 {
			a.prevScreen = ScreenPlaylistDetail
			a.screen = ScreenNowPlaying
			return a, a.playFromList(songs, a.playlistDetCursor)
		}
	case "p":
		a.pl.Pause()
	case "r":
		if len(songs) > 0 {
			if _, err := a.lib.RemoveFromPlaylist(a.activePlaylist.ID, songs[a.playlistDetCursor].ID); err != nil {
				return a, a.setStatus("Save error: "+err.Error(), true)
			}
			remaining := a.lib.GetPlaylistSongs(a.activePlaylist.ID)
			if a.playlistDetCursor >= len(remaining) && a.playlistDetCursor > 0 {
				a.playlistDetCursor--
			}
			return a, a.setStatus("Removed from playlist", false)
		}
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenPlaylistDetail
			a.screen = ScreenNowPlaying
		}
	case "esc", "h":
		a.screen = ScreenPlaylists
	}
	return a, nil
}

// ── Now Playing ───────────────────────────────────────────────────────────────

func (a *App) updateNowPlaying(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "p", " ":
		a.pl.Pause()
	case "s":
		a.pl.Stop()
		a.nowPlayingSong = nil
	case ">":
		return a, a.advanceQueue()
	case "<":
		return a, a.rewindQueue()
	case "plus", "=":
		a.pl.VolumeUp()
	case "minus":
		a.pl.VolumeDown()
	case "left":
		a.pl.Seek(-10 * time.Second)
	case "right":
		a.pl.Seek(10 * time.Second)
	case "esc", "h":
		// go back — never recurse into NowPlaying
		back := a.prevScreen
		if back == ScreenNowPlaying || back == ScreenNone {
			back = ScreenHome
		}
		a.screen = back
	case "l":
		a.screen = ScreenLibrary
	}
	return a, nil
}

// ── Add Song Form ─────────────────────────────────────────────────────────────

func (a *App) updateAddSong(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		a.screen = ScreenHome
		return a, nil
	case "tab", "down":
		a.addSongFields[a.addSongFocus].Blur()
		a.addSongFocus = (a.addSongFocus + 1) % len(a.addSongFields)
		a.addSongFields[a.addSongFocus].Focus()
		return a, nil
	case "shift+tab", "up":
		a.addSongFields[a.addSongFocus].Blur()
		a.addSongFocus = (a.addSongFocus - 1 + len(a.addSongFields)) % len(a.addSongFields)
		a.addSongFields[a.addSongFocus].Focus()
		return a, nil
	case "enter":
		if a.addSongFocus < len(a.addSongFields)-1 {
			a.addSongFields[a.addSongFocus].Blur()
			a.addSongFocus++
			a.addSongFields[a.addSongFocus].Focus()
			return a, nil
		}
		// submit
		title := strings.TrimSpace(a.addSongFields[0].Value())
		artist := strings.TrimSpace(a.addSongFields[1].Value())
		album := strings.TrimSpace(a.addSongFields[2].Value())
		genre := strings.TrimSpace(a.addSongFields[3].Value())
		fp := strings.TrimSpace(a.addSongFields[4].Value())
		durStr := strings.TrimSpace(a.addSongFields[5].Value())
		if title == "" || fp == "" {
			return a, a.setStatus("Title and File Path are required", true)
		}
		dur := 0
		if durStr != "" {
			var err error
			dur, err = strconv.Atoi(durStr)
			if err != nil {
				return a, a.setStatus("Duration must be a number in seconds", true)
			}
		} else {
			// Auto-detect duration from file
			if detected, err := library.DetectDuration(fp); err == nil && detected > 0 {
				dur = detected
			}
		}
		if _, err := a.lib.AddSong(title, artist, album, genre, fp, dur); err != nil {
			return a, a.setStatus("Save error: "+err.Error(), true)
		}
		cmd := a.setStatus("Song added: "+title, false)
		a.screen = ScreenLibrary
		a.libCursor = len(a.lib.Songs) - 1
		return a, cmd
	}
	var cmd tea.Cmd
	a.addSongFields[a.addSongFocus], cmd = a.addSongFields[a.addSongFocus].Update(msg)
	return a, cmd
}

func (a *App) resetAddSongForm() {
	for i := range a.addSongFields {
		a.addSongFields[i].SetValue("")
		a.addSongFields[i].Blur()
	}
	a.addSongFocus = 0
	a.addSongFields[0].Focus()
}

// ── Create Playlist ───────────────────────────────────────────────────────────

func (a *App) updateCreatePlaylist(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		a.screen = ScreenPlaylists
		return a, nil
	case "enter":
		name := strings.TrimSpace(a.createPLInput.Value())
		if name == "" {
			return a, a.setStatus("Playlist name cannot be empty", true)
		}
		if _, err := a.lib.CreatePlaylist(name); err != nil {
			return a, a.setStatus("Save error: "+err.Error(), true)
		}
		cmd := a.setStatus("Playlist created: "+name, false)
		a.screen = ScreenPlaylists
		a.playlistCursor = len(a.lib.Playlists) - 1
		return a, cmd
	}
	var cmd tea.Cmd
	a.createPLInput, cmd = a.createPLInput.Update(msg)
	return a, cmd
}

// ── Add to Playlist selection ─────────────────────────────────────────────────

func (a *App) updateAddToPlaylist(key string) (tea.Model, tea.Cmd) {
	pls := a.lib.Playlists
	switch key {
	case "up", "k":
		if a.addToPLCursor > 0 {
			a.addToPLCursor--
		}
	case "down", "j":
		if a.addToPLCursor < len(pls)-1 {
			a.addToPLCursor++
		}
	case "enter":
		if len(pls) > 0 && a.addToPLCursor < len(pls) {
			ok, err := a.lib.AddToPlaylist(pls[a.addToPLCursor].ID, a.addToPLSongID)
			if err != nil {
				return a, a.setStatus("Save error: "+err.Error(), true)
			}
			a.screen = a.prevScreen
			if ok {
				return a, a.setStatus("Added to "+pls[a.addToPLCursor].Name, false)
			}
			return a, a.setStatus("Song already in playlist", true)
		}
	case "esc", "h":
		a.screen = a.prevScreen
	}
	return a, nil
}

// ── Queue helpers ─────────────────────────────────────────────────────────────

func (a *App) playFromList(songs []library.Song, idx int) tea.Cmd {
	if idx < 0 || idx >= len(songs) {
		return nil
	}
	a.queueSongs = make([]library.Song, len(songs))
	copy(a.queueSongs, songs)
	a.queueIndex = idx
	return a.playCurrent()
}

func (a *App) playCurrent() tea.Cmd {
	if a.queueIndex < 0 || a.queueIndex >= len(a.queueSongs) {
		return nil
	}
	song := a.queueSongs[a.queueIndex]

	// Detect if this is a saved online track (URL instead of local file)
	if isURL(song.FilePath) {
		a.onlineLoading = true
		s := song
		return func() tea.Msg {
			return reDownloadSavedTrack(s)
		}
	}

	if err := validateAudioPath(song.FilePath); err != nil {
		return a.setStatus("Invalid file: "+err.Error(), true)
	}
	if err := a.pl.Play(song.FilePath); err != nil {
		cmd := a.setStatus("Playback error: "+err.Error(), true)
		a.nowPlayingSong = nil
		return cmd
	}
	cp := song
	a.nowPlayingSong = &cp
	return tickCmd()
}

// isURL checks if a path is a web URL
func isURL(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// reDownloadSavedTrack downloads a saved online track via yt-dlp
func reDownloadSavedTrack(song library.Song) savedTrackReadyMsg {
	path, err := online.YtdlpDownloadTemp(song.FilePath)
	if err != nil {
		return savedTrackReadyMsg{err: err}
	}
	return savedTrackReadyMsg{tempPath: path, song: song}
}

// validateAudioPath checks that a file path is a local file with a valid
// audio extension. Rejects network paths, URLs, and missing files.
func validateAudioPath(path string) error {
	// Reject URLs and network paths
	lower := strings.ToLower(path)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "\\\\") || strings.HasPrefix(lower, "//") ||
		strings.Contains(lower, "://") {
		return &os.PathError{Op: "open", Path: path, Err: os.ErrInvalid}
	}
	// Check extension
	validExt := map[string]bool{".mp3": true, ".wav": true, ".flac": true}
	if !validExt[filepath.Ext(lower)] {
		return &os.PathError{Op: "open", Path: path, Err: os.ErrInvalid}
	}
	// Check file exists
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func (a *App) advanceQueue() tea.Cmd {
	if a.queueIndex < len(a.queueSongs)-1 {
		a.queueIndex++
		return a.playCurrent()
	}
	a.pl.Stop()
	a.nowPlayingSong = nil
	return nil
}

func (a *App) rewindQueue() tea.Cmd {
	if a.queueIndex > 0 {
		a.queueIndex--
		return a.playCurrent()
	}
	return nil
}

// ── Online Search (source picker + query input) ───────────────────────────────

func (a *App) updateOnlineSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Auto-focus input when entering the screen
	if !a.onlineInput.Focused() {
		switch key {
		case "enter", "/":
			a.onlineInput.Focus()
			return a, nil
		case "esc", "h":
			a.screen = ScreenHome
			return a, nil
		}
		return a, nil
	}

	// Input is focused
	switch key {
	case "esc":
		a.onlineInput.Blur()
		return a, nil
	case "enter":
		q := strings.TrimSpace(a.onlineInput.Value())
		if q == "" {
			return a, nil
		}
		a.onlineLoading = true
		a.onlineErr = ""
		a.onlineResults = nil
		a.screen = ScreenOnlineResults
		return a, func() tea.Msg {
			return doOnlineSearch(q)
		}
	default:
		var cmd tea.Cmd
		a.onlineInput, cmd = a.onlineInput.Update(msg)
		return a, cmd
	}
}

// doOnlineSearch runs in a goroutine and returns onlineResultsMsg
func doOnlineSearch(query string) tea.Msg {
	tracks, err := online.YtdlpSearch(query, 10)
	if err != nil {
		return onlineResultsMsg{err: err}
	}
	return onlineResultsMsg{tracks: tracks}
}

// ── Online Results ────────────────────────────────────────────────────────────

func (a *App) updateOnlineResults(key string) (tea.Model, tea.Cmd) {
	if a.onlineLoading {
		if key == "esc" || key == "h" {
			a.onlineLoading = false
			a.screen = ScreenOnlineSearch
		}
		return a, nil
	}

	switch key {
	case "up", "k":
		if a.onlineCursor > 0 {
			a.onlineCursor--
		}
	case "down", "j":
		if a.onlineCursor < len(a.onlineResults)-1 {
			a.onlineCursor++
		}
	case "enter":
		if len(a.onlineResults) == 0 || a.onlineCursor >= len(a.onlineResults) {
			return a, nil
		}
		track := a.onlineResults[a.onlineCursor]
		a.onlineLoading = true
		a.prevScreen = ScreenOnlineResults
		return a, func() tea.Msg {
			return playOnlineTrack(track)
		}
	case "s":
		if len(a.onlineResults) > 0 && a.onlineCursor < len(a.onlineResults) {
			t := a.onlineResults[a.onlineCursor]
			if _, err := a.lib.AddSong(t.Title, t.Artist, t.Album, "Online / "+t.Source, t.StreamURL, t.Duration); err != nil {
				return a, a.setStatus("Save error: "+err.Error(), true)
			}
			return a, a.setStatus("Saved to library: "+t.Title, false)
		}
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenOnlineResults
			a.screen = ScreenNowPlaying
		}
	case "esc", "h":
		a.onlineCursor = 0
		a.screen = ScreenOnlineSearch
	}
	return a, nil
}

// playOnlineTrack downloads via yt-dlp and returns an onlinePlayReadyMsg
func playOnlineTrack(track online.OnlineTrack) tea.Msg {
	path, err := online.YtdlpDownloadTemp(track.StreamURL)
	return onlinePlayReadyMsg{tempPath: path, track: track, err: err}
}

// ── Recommended ───────────────────────────────────────────────────────────────

func (a *App) updateRecommended(key string) (tea.Model, tea.Cmd) {
	genres := online.Genres

	// ── Results view ──────────────────────────────────────────────────────────
	if a.recActiveGenre != nil {
		if a.recLoading {
			if key == "esc" || key == "h" {
				a.recLoading = false
				a.recActiveGenre = nil
			}
			return a, nil
		}
		switch key {
		case "up", "k":
			if a.recResCursor > 0 {
				a.recResCursor--
			}
		case "down", "j":
			if a.recResCursor < len(a.recResults)-1 {
				a.recResCursor++
			}
		case "enter":
			if len(a.recResults) > 0 && a.recResCursor < len(a.recResults) {
				track := a.recResults[a.recResCursor]
				a.recLoading = true
				a.prevScreen = ScreenRecommended
				return a, func() tea.Msg {
					return playOnlineTrack(track)
				}
			}
		case "s":
			if len(a.recResults) > 0 && a.recResCursor < len(a.recResults) {
				t := a.recResults[a.recResCursor]
				if _, err := a.lib.AddSong(t.Title, t.Artist, t.Album, a.recActiveGenre.Name, t.StreamURL, t.Duration); err != nil {
					return a, a.setStatus("Save error: "+err.Error(), true)
				}
				return a, a.setStatus("Saved: "+t.Title, false)
			}
		case "b", "esc":
			a.recActiveGenre = nil
			a.recResults = nil
			a.recErr = ""
		case "n":
			if a.nowPlayingSong != nil {
				a.prevScreen = ScreenRecommended
				a.screen = ScreenNowPlaying
			}
		case "h":
			a.screen = ScreenHome
		}
		return a, nil
	}

	// ── Genre picker ──────────────────────────────────────────────────────────
	switch key {
	case "up", "k":
		if a.recCursor > 0 {
			a.recCursor--
		}
	case "down", "j":
		if a.recCursor < len(genres)-1 {
			a.recCursor++
		}
	case "enter":
		if a.recCursor < 0 || a.recCursor >= len(genres) {
			return a, nil
		}
		g := genres[a.recCursor]
		a.recActiveGenre = &online.Genres[a.recCursor]
		a.recLoading = true
		a.recErr = ""
		a.recResults = nil
		a.recResCursor = 0
		return a, func() tea.Msg {
			tracks, err := online.RecommendedSearch(g)
			return recResultsMsg{tracks: tracks, err: err}
		}
	case "n":
		if a.nowPlayingSong != nil {
			a.prevScreen = ScreenRecommended
			a.screen = ScreenNowPlaying
		}
	case "esc", "h":
		a.screen = ScreenHome
	}
	return a, nil
}
