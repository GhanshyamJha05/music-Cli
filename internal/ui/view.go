package ui

import (
	"fmt"
	"github.com/GhanshyamJha05/music-Cli/internal/library"
	"github.com/GhanshyamJha05/music-Cli/internal/online"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) View() string {
	var b strings.Builder

	switch a.screen {
	case ScreenHome:
		b.WriteString(a.viewHome())
	case ScreenLibrary:
		b.WriteString(a.viewLibrary())
	case ScreenSearch:
		b.WriteString(a.viewSearch())
	case ScreenPlaylists:
		b.WriteString(a.viewPlaylists())
	case ScreenPlaylistDetail:
		b.WriteString(a.viewPlaylistDetail())
	case ScreenNowPlaying:
		b.WriteString(a.viewNowPlaying())
	case ScreenAddSong:
		b.WriteString(a.viewAddSong())
	case ScreenCreatePlaylist:
		b.WriteString(a.viewCreatePlaylist())
	case ScreenAddToPlaylist:
		b.WriteString(a.viewAddToPlaylist())
	case ScreenOnlineSearch:
		b.WriteString(a.viewOnlineSearch())
	case ScreenOnlineResults:
		b.WriteString(a.viewOnlineResults())
	case ScreenRecommended:
		b.WriteString(a.viewRecommended())
	}

	// Sticky Mini-player for screens other than Home and NowPlaying
	if a.screen != ScreenHome && a.screen != ScreenNowPlaying && a.nowPlayingSong != nil {
		b.WriteString("\n" + a.miniNowPlaying() + "\n")
	}

	// Status bar at bottom
	if a.statusMsg != "" {
		var s string
		if a.statusIsError {
			s = errorStyle.Render("✗ " + a.statusMsg)
		} else {
			s = successStyle.Render("✓ " + a.statusMsg)
		}
		b.WriteString("\n" + s)
	}

	return b.String()
}

// ── Home ──────────────────────────────────────────────────────────────────────

func (a *App) viewHome() string {
	var b strings.Builder

	b.WriteString(Logo())
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("  Your personal music player — right from the terminal") + "\n\n")

	menu := []struct{ key, label, desc string }{
		{"1", "Library", "Browse & play all your songs"},
		{"2", "Search", "Find songs by title, artist or genre"},
		{"3", "Playlists", "Manage your custom playlists"},
		{"4", "Add Song", "Add a new song to your library"},
		{"5", "Online Browse", "Search YouTube and stream"},
		{"6", "Recommended", "Curated picks by genre — Pop, Lo-Fi, Funk & more"},
	}

	for _, m := range menu {
		line := "  " + keyHintStyle.Render("["+m.key+"]") + "  " +
			lipgloss.NewStyle().Foreground(white).Bold(true).Render(m.label) +
			"  " + keyDescStyle.Render(m.desc)
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")

	// Now playing mini bar
	if a.nowPlayingSong != nil {
		b.WriteString(a.miniNowPlaying() + "\n")
		b.WriteString("  " + keyHintStyle.Render("[n]") + " " + keyDescStyle.Render("Open Now Playing") + "\n")
	}

	b.WriteString("\n  " + keyHintStyle.Render("[q]") + " " + keyDescStyle.Render("Quit"))
	return b.String()
}

func (a *App) miniNowPlaying() string {
	if a.nowPlayingSong == nil {
		return ""
	}
	state := "▶"
	badge := playingBadge
	if a.pl.IsPaused() {
		state = "⏸"
		badge = pausedBadge
	}

	elapsed := a.pl.Elapsed()
	elapsedStr := fmt.Sprintf("%d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	totalStr := library.FormatDuration(a.nowPlayingSong.Duration)

	loopBadge := ""
	switch a.loopMode {
	case LoopAll:
		loopBadge = "  " + lipgloss.NewStyle().Foreground(green).Bold(true).Render("🔁 ALL")
	case LoopOne:
		loopBadge = "  " + lipgloss.NewStyle().Foreground(yellow).Bold(true).Render("🔂 ONE")
	}

	vol := int(a.pl.GetVolume() * 100)
	volStr := dimItemStyle.Render(fmt.Sprintf("Vol: %d%%", vol))

	songInfo := lipgloss.NewStyle().Foreground(white).Bold(true).Render(truncate(a.nowPlayingSong.Title, 30)) +
		"  " + subtitleStyle.Render(truncate(a.nowPlayingSong.Artist, 20))

	timeInfo := lipgloss.NewStyle().Foreground(lightBlue).Render(fmt.Sprintf("[%s / %s]", elapsedStr, totalStr))

	return nowPlayingStyle.Render(
		badge.Render(state) + "  " + songInfo + "   " + timeInfo + "   " + volStr + loopBadge,
	)
}

// ── Library ───────────────────────────────────────────────────────────────────

func (a *App) viewLibrary() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("♫  Library") + "\n")
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d songs", len(a.lib.Songs))) + "\n\n")

	if len(a.lib.Songs) == 0 {
		b.WriteString(dimItemStyle.Render("  No songs yet. Press [4] to add one.") + "\n")
	} else {
		// reserved: title(1) + subtitle(1) + blank(1) + header(2) + blank(1) + hints(1) + status(0-1)
		reserved := 8
		if a.statusMsg != "" {
			reserved++
		}
		if a.nowPlayingSong != nil {
			reserved += 2
		}
		b.WriteString(a.renderSongList(a.lib.Songs, a.libCursor, a.nowPlayingSong, reserved))
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"↑/↓", "navigate"}, {"enter", "play"}, {"p", "pause"},
		{"s", "stop"}, {"r", "loop"}, {"tab/>", "skip"},
		{"K/J", "reorder"}, {"a", "add to PL"}, {"d", "remove"},
		{"n", "now playing"}, {"h", "back"},
	}))
	return b.String()
}

// ── Search ────────────────────────────────────────────────────────────────────

func (a *App) viewSearch() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🔍  Search") + "\n\n")

	searchBox := inputStyle.Render("  " + a.searchInput.View())
	b.WriteString(searchBox + "\n")
	if a.searchActive {
		b.WriteString(dimItemStyle.Render("  Type to search, Enter to confirm, Esc to cancel") + "\n")
	} else {
		b.WriteString(dimItemStyle.Render("  Press / to search again") + "\n")
	}
	b.WriteString("\n")

	if len(a.searchResults) == 0 && !a.searchActive {
		b.WriteString(dimItemStyle.Render("  No results.") + "\n")
	} else if len(a.searchResults) > 0 {
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d results", len(a.searchResults))) + "\n\n")
		// reserved: title(1) + blank(1) + searchBox(1) + hintLine(1) + blank(1) + subtitle(1) + blank(1) + header(2) + blank(1) + hints(1) + status(0-1)
		reserved := 11
		if a.statusMsg != "" {
			reserved++
		}
		if a.nowPlayingSong != nil {
			reserved += 2
		}
		b.WriteString(a.renderSongList(a.searchResults, a.searchCursor, a.nowPlayingSong, reserved))
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"/", "search"}, {"↑/↓", "navigate"}, {"enter", "play"},
		{"n", "now playing"}, {"h", "back"},
	}))
	return b.String()
}

// ── Playlists ─────────────────────────────────────────────────────────────────

func (a *App) viewPlaylists() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("≡  Playlists") + "\n")
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d playlists", len(a.lib.Playlists))) + "\n\n")

	if len(a.lib.Playlists) == 0 {
		b.WriteString(dimItemStyle.Render("  No playlists. Press [c] to create one.") + "\n")
	} else {
		for i, pl := range a.lib.Playlists {
			count := fmt.Sprintf("%d songs", len(pl.SongIDs))
			line := fmt.Sprintf("  %-30s %s", pl.Name, dimItemStyle.Render(count))
			if i == a.playlistCursor {
				b.WriteString(selectedItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(normalItemStyle.Render(line) + "\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"↑/↓", "navigate"}, {"enter", "open"}, {"c", "create"},
		{"d", "delete"}, {"h", "back"},
	}))
	return b.String()
}

// ── Playlist Detail ───────────────────────────────────────────────────────────

func (a *App) viewPlaylistDetail() string {
	var b strings.Builder
	if a.activePlaylist == nil {
		b.WriteString(dimItemStyle.Render("  No playlist selected.") + "\n")
		b.WriteString(renderHints([]hint{{"h", "back"}}))
		return b.String()
	}
	name := a.activePlaylist.Name
	b.WriteString(titleStyle.Render("≡  "+name) + "\n")
	songs := a.lib.GetPlaylistSongs(a.activePlaylist.ID)
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d songs", len(songs))) + "\n\n")

	if len(songs) == 0 {
		b.WriteString(dimItemStyle.Render("  No songs. Go to Library and press [a] to add.") + "\n")
	} else {
		// reserved: title(1) + subtitle(1) + blank(1) + header(2) + blank(1) + hints(1) + status(0-1)
		reserved := 8
		if a.statusMsg != "" {
			reserved++
		}
		if a.nowPlayingSong != nil {
			reserved += 2
		}
		b.WriteString(a.renderSongList(songs, a.playlistDetCursor, a.nowPlayingSong, reserved))
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"↑/↓", "navigate"}, {"enter", "play"}, {"p", "pause"},
		{"s", "stop"}, {"+/-", "volume"}, {"tab/>", "skip"},
		{"K/J", "reorder"}, {"d/r", "remove"}, {"n", "now playing"}, {"esc/h", "back"},
	}))
	return b.String()
}

// ── Now Playing ───────────────────────────────────────────────────────────────

func (a *App) viewNowPlaying() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("♫  Now Playing") + "\n\n")

	if a.nowPlayingSong == nil {
		if a.onlineLoading {
			spinner := lipgloss.NewStyle().Foreground(green).Bold(true).Render("⠋ ")
			b.WriteString("\n  " + spinner + warningStyle.Render("Re-downloading track via yt-dlp...") + "\n")
			b.WriteString(dimItemStyle.Render("  This takes a few seconds") + "\n")
			b.WriteString("\n" + renderHints([]hint{{"esc", "cancel"}}))
			return b.String()
		}
		b.WriteString(dimItemStyle.Render("  Nothing playing.") + "\n")
		b.WriteString(renderHints([]hint{{"h", "back"}, {"l", "library"}}))
		return b.String()
	}

	song := a.nowPlayingSong

	// Album art placeholder (ascii block)
	art := lipgloss.NewStyle().
		Foreground(green).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(darkGreen).
		Padding(1, 3).
		Render("♪\n♫\n♪")

	info := strings.Builder{}
	info.WriteString(lipgloss.NewStyle().Foreground(white).Bold(true).Render(song.Title) + "\n")
	info.WriteString(subtitleStyle.Render(song.Artist) + "\n")
	info.WriteString(dimItemStyle.Render(song.Album) + "\n")
	if song.Genre != "" {
		info.WriteString(badgeStyle.Render(song.Genre) + "\n")
	}

	// Progress
	elapsed := a.pl.Elapsed()
	var pct float64
	if song.Duration > 0 {
		pct = elapsed.Seconds() / float64(song.Duration)
		if pct > 1 {
			pct = 1
		}
	}
	elapsedStr := fmt.Sprintf("%d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	totalStr := library.FormatDuration(song.Duration)

	stateStr := playingBadge.Render("▶ PLAYING")
	if a.pl.IsPaused() {
		stateStr = pausedBadge.Render("⏸ PAUSED")
	} else if !a.pl.IsPlaying() {
		stateStr = badgeStyle.Render("■ STOPPED")
	}

	progressLine := "  " + elapsedStr + "  " + ProgressBar(pct, 35) + "  " + totalStr

	// Volume bar
	vol := a.pl.GetVolume()
	volPct := int(vol * 100)
	volBar := "  " + keyHintStyle.Render("Vol") + "  " + VolumeBar(vol, 20) + "  " + dimItemStyle.Render(fmt.Sprintf("%d%%", volPct))

	row := lipgloss.JoinHorizontal(lipgloss.Top, art, "   ", info.String())
	b.WriteString(boxStyle.Render(row) + "\n\n")
	b.WriteString("  " + stateStr + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lightBlue).Render(progressLine) + "\n\n")
	b.WriteString(volBar + "\n\n")

	// Queue info & Loop mode
	var queueParts []string
	if len(a.queueSongs) > 1 {
		queueParts = append(queueParts, fmt.Sprintf("Track %d / %d", a.queueIndex+1, len(a.queueSongs)))
	}
	switch a.loopMode {
	case LoopAll:
		queueParts = append(queueParts, lipgloss.NewStyle().Foreground(green).Bold(true).Render("🔁 LOOP ALL"))
	case LoopOne:
		queueParts = append(queueParts, lipgloss.NewStyle().Foreground(yellow).Bold(true).Render("🔂 LOOP ONE"))
	}
	if len(queueParts) > 0 {
		b.WriteString("  " + dimItemStyle.Render(strings.Join(queueParts, "  •  ")) + "\n\n")
	}

	b.WriteString(renderHints([]hint{
		{"space/p", "pause/resume"}, {"s", "stop"}, {">", "next"}, {"<", "prev"},
	}))
	b.WriteString(renderHints([]hint{
		{"r", "loop (off/all/one)"}, {"←/→", "±5s seek"}, {"[/]", "±15s"}, {"+/-", "vol"}, {"esc/h", "back"},
	}))
	return b.String()
}

// ── Add Song ──────────────────────────────────────────────────────────────────

func (a *App) viewAddSong() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("+ Add Song") + "\n")
	b.WriteString(subtitleStyle.Render("  Fill in the details (Tab to move between fields)") + "\n\n")

	labels := []string{"Title *", "Artist", "Album", "Genre", "File Path *", "Duration (auto if blank)"}
	for i, f := range a.addSongFields {
		label := labels[i]
		if i == a.addSongFocus {
			b.WriteString("  " + keyHintStyle.Render("› "+label) + "\n")
		} else {
			b.WriteString("  " + dimItemStyle.Render("  "+label) + "\n")
		}
		b.WriteString("  " + inputStyle.Render(f.View()) + "\n\n")
	}

	b.WriteString(renderHints([]hint{
		{"tab/↓", "next field"}, {"shift+tab/↑", "prev field"},
		{"enter", "submit"}, {"esc", "cancel"},
	}))
	return b.String()
}

// ── Create Playlist ───────────────────────────────────────────────────────────

func (a *App) viewCreatePlaylist() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("+ Create Playlist") + "\n\n")
	b.WriteString("  " + keyHintStyle.Render("Name:") + "\n")
	b.WriteString("  " + inputStyle.Render(a.createPLInput.View()) + "\n\n")
	b.WriteString(renderHints([]hint{{"enter", "create"}, {"esc", "cancel"}}))
	return b.String()
}

// ── Add to Playlist ───────────────────────────────────────────────────────────

func (a *App) viewAddToPlaylist() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Add to Playlist") + "\n\n")

	if len(a.lib.Playlists) == 0 {
		b.WriteString(dimItemStyle.Render("  No playlists. Create one first (Playlists → c).") + "\n")
	} else {
		for i, pl := range a.lib.Playlists {
			count := fmt.Sprintf("%d songs", len(pl.SongIDs))
			line := fmt.Sprintf("  %-30s %s", pl.Name, dimItemStyle.Render(count))
			if i == a.addToPLCursor {
				b.WriteString(selectedItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(normalItemStyle.Render(line) + "\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"↑/↓", "navigate"}, {"enter", "select"}, {"esc", "cancel"},
	}))
	return b.String()
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// renderSongList renders a scrollable song list with cursor, dynamically adapting
// to available terminal width and padding the selected line cleanly.
func (a *App) renderSongList(songs []library.Song, cursor int, nowPlaying *library.Song, reservedLines int) string {
	var b strings.Builder

	termWidth := a.width
	if termWidth <= 0 {
		termWidth = 80
	}

	// Layout widths
	numW := 4
	durW := 6
	markW := 2

	// Flexible columns: Title, Artist, Album
	remain := termWidth - (markW + numW + durW + 8)
	if remain < 30 {
		remain = 30
	}

	titleW := int(float64(remain) * 0.45)
	artistW := int(float64(remain) * 0.30)
	albumW := remain - titleW - artistW
	if titleW < 14 {
		titleW = 14
	}
	if artistW < 10 {
		artistW = 10
	}
	if albumW < 6 {
		albumW = 6
	}

	header := fmt.Sprintf("  %-*s %-*s %-*s %*s",
		titleW, "TITLE", artistW, "ARTIST", albumW, "ALBUM", durW, "TIME")
	b.WriteString(lipgloss.NewStyle().Foreground(gray).Bold(true).Render(header) + "\n")

	dividerW := termWidth
	if dividerW > 120 {
		dividerW = 120
	}
	b.WriteString(lipgloss.NewStyle().Foreground(darkGray).Render(strings.Repeat("─", dividerW)) + "\n")

	if len(songs) == 0 {
		return b.String()
	}

	height := a.height
	if height <= 0 {
		height = 24
	}
	visibleRows := height - reservedLines
	if visibleRows < 1 {
		visibleRows = 1
	}

	start := cursor - visibleRows/2
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > len(songs) {
		end = len(songs)
	}
	if end-start < visibleRows && start > 0 {
		start = end - visibleRows
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		s := songs[i]
		num := fmt.Sprintf("%2d.", i+1)
		title := truncate(s.Title, titleW-1)
		artist := truncate(s.Artist, artistW-1)
		album := truncate(s.Album, albumW-1)
		dur := library.FormatDuration(s.Duration)

		isPlaying := nowPlaying != nil && nowPlaying.ID == s.ID
		playMark := "  "
		if isPlaying {
			playMark = "♫ "
		}

		line := fmt.Sprintf("%s%-3s %-*s %-*s %-*s %*s",
			playMark, num, titleW, title, artistW, artist, albumW, album, durW, dur)

		if i == cursor {
			b.WriteString(selectedItemStyle.Width(dividerW).Render(line) + "\n")
		} else if isPlaying {
			b.WriteString(lipgloss.NewStyle().Foreground(accent).Bold(true).Render(line) + "\n")
		} else {
			b.WriteString(normalItemStyle.Render(line) + "\n")
		}
	}
	return b.String()
}

type hint struct{ key, desc string }

func renderHints(hints []hint) string {
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = keyHintStyle.Render("["+h.key+"]") + " " + keyDescStyle.Render(h.desc)
	}
	return "  " + strings.Join(parts, "  ") + "\n"
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max-1]) + "…"
}

// ── Online Search ─────────────────────────────────────────────────────────────

func (a *App) viewOnlineSearch() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🌐  Online Browse") + "\n")
	b.WriteString(subtitleStyle.Render("  Stream from YouTube via yt-dlp") + "\n\n")

	// Search input
	if a.onlineInput.Focused() {
		b.WriteString(keyHintStyle.Render("  Query:") + "\n")
		b.WriteString("  " + inputStyle.Render(a.onlineInput.View()) + "\n")
		b.WriteString(dimItemStyle.Render("  Enter to search · Esc to cancel") + "\n")
	} else {
		b.WriteString(dimItemStyle.Render("  Press Enter or / to type your search query") + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"enter//", "type query"}, {"esc/h", "back"},
	}))
	return b.String()
}

// ── Online Results ────────────────────────────────────────────────────────────

func (a *App) viewOnlineResults() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🌐  Online Results") + "\n")

	if a.onlineLoading {
		spinner := lipgloss.NewStyle().Foreground(green).Bold(true).Render("⠋ ")
		b.WriteString("\n  " + spinner + warningStyle.Render("Searching / downloading... please wait") + "\n")
		b.WriteString(dimItemStyle.Render("  (yt-dlp downloads may take 10–30 seconds)") + "\n")
		b.WriteString("\n" + renderHints([]hint{{"esc", "cancel"}}))
		return b.String()
	}

	if a.onlineErr != "" {
		b.WriteString("\n  " + errorStyle.Render("✗ "+a.onlineErr) + "\n")
		b.WriteString(renderHints([]hint{{"h", "back"}}))
		return b.String()
	}

	if len(a.onlineResults) == 0 {
		b.WriteString(dimItemStyle.Render("  No results found.") + "\n")
		b.WriteString(renderHints([]hint{{"h", "back"}}))
		return b.String()
	}

	b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d tracks found", len(a.onlineResults))) + "\n\n")

	// Table header
	header := fmt.Sprintf("  %-4s %-26s %-20s %-12s %6s  %-8s",
		"#", "TITLE", "ARTIST", "ALBUM", "TIME", "SOURCE")
	b.WriteString(lipgloss.NewStyle().Foreground(gray).Bold(true).Render(header) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(darkGray).Render(strings.Repeat("─", 82)) + "\n")

	// Virtual scrolling: calculate visible window
	// reserved: title(1) + subtitle(1) + blank(1) + header(1) + separator(1) + blank(1) + hints(1) + status(0-1)
	reserved := 7
	if a.statusMsg != "" {
		reserved++
	}
	visibleRows := a.height - reserved
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := a.onlineCursor - visibleRows/2
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > len(a.onlineResults) {
		end = len(a.onlineResults)
	}
	if end-start < visibleRows && start > 0 {
		start = end - visibleRows
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		t := a.onlineResults[i]
		num := fmt.Sprintf("%3d", i+1)
		title := truncate(t.Title, 25)
		artist := truncate(t.Artist, 19)
		album := truncate(t.Album, 11)
		dur := library.FormatDuration(t.Duration)

		srcBadge := lipgloss.NewStyle().Foreground(red).Render("youtube")

		isPlaying := a.nowPlayingSong != nil && a.nowPlayingSong.Title == t.Title && a.nowPlayingSong.Artist == t.Artist
		playMark := "  "
		if isPlaying {
			playMark = lipgloss.NewStyle().Foreground(green).Render("♫ ")
		}

		line := fmt.Sprintf("%s%-4s %-26s %-20s %-12s %6s  %s",
			playMark, num, title, artist, album, dur, srcBadge)

		if i == a.onlineCursor {
			b.WriteString(selectedItemStyle.Render(line) + "\n")
		} else {
			b.WriteString(normalItemStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"↑/↓", "navigate"}, {"enter", "stream & play"},
		{"s/a", "add to library"}, {"n", "now playing"}, {"h", "back"},
	}))
	return b.String()
}

// ── Recommended ───────────────────────────────────────────────────────────────

func (a *App) viewRecommended() string {
	var b strings.Builder

	// ── Results view (genre selected) ────────────────────────────────────────
	if a.recActiveGenre != nil {
		genre := a.recActiveGenre

		// header with genre colour
		genreColor := lipgloss.Color(genre.Color)
		title := lipgloss.NewStyle().Bold(true).Foreground(genreColor).Padding(0, 1).
			Render(genre.Emoji + "  " + genre.Name)
		b.WriteString(title + "\n")

		if a.recLoading {
			spinner := lipgloss.NewStyle().Foreground(genreColor).Bold(true).Render("⠋ ")
			b.WriteString("\n  " + spinner + warningStyle.Render("Fetching "+genre.Name+" picks via yt-dlp...") + "\n")
			b.WriteString(dimItemStyle.Render("  This takes a few seconds") + "\n")
			b.WriteString("\n" + renderHints([]hint{{"esc", "cancel"}}))
			return b.String()
		}

		if a.recErr != "" {
			b.WriteString("\n  " + errorStyle.Render("✗ "+a.recErr) + "\n")
			b.WriteString(renderHints([]hint{{"b/esc", "back to genres"}, {"h", "home"}}))
			return b.String()
		}

		if len(a.recResults) == 0 {
			b.WriteString(dimItemStyle.Render("  No results found.") + "\n")
			b.WriteString(renderHints([]hint{{"b/esc", "back to genres"}}))
			return b.String()
		}

		b.WriteString(subtitleStyle.Render(fmt.Sprintf("  %d tracks", len(a.recResults))) + "\n\n")

		// table header
		header := fmt.Sprintf("  %-4s %-32s %-24s %6s",
			"#", "TITLE", "ARTIST / CHANNEL", "TIME")
		b.WriteString(lipgloss.NewStyle().Foreground(gray).Bold(true).Render(header) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(darkGray).Render(strings.Repeat("─", 72)) + "\n")

		// Virtual scrolling: calculate visible window
		// reserved: title(1) + subtitle(1) + blank(1) + header(1) + separator(1) + blank(1) + hints(1) + status(0-1)
		reserved := 7
		if a.statusMsg != "" {
			reserved++
		}
		visibleRows := a.height - reserved
		if visibleRows < 1 {
			visibleRows = 1
		}
		start := a.recResCursor - visibleRows/2
		if start < 0 {
			start = 0
		}
		end := start + visibleRows
		if end > len(a.recResults) {
			end = len(a.recResults)
		}
		if end-start < visibleRows && start > 0 {
			start = end - visibleRows
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			t := a.recResults[i]
			num := fmt.Sprintf("%3d", i+1)
			title := truncate(t.Title, 31)
			artist := truncate(t.Artist, 23)
			dur := library.FormatDuration(t.Duration)

			isPlaying := a.nowPlayingSong != nil &&
				a.nowPlayingSong.Title == t.Title && a.nowPlayingSong.Artist == t.Artist
			playMark := "  "
			if isPlaying {
				playMark = lipgloss.NewStyle().Foreground(lipgloss.Color(genre.Color)).Render("♫ ")
			}

			line := fmt.Sprintf("%s%-4s %-32s %-24s %6s",
				playMark, num, title, artist, dur)

			if i == a.recResCursor {
				b.WriteString(lipgloss.NewStyle().
					Foreground(black).Background(lipgloss.Color(genre.Color)).
					Bold(true).Padding(0, 1).Render(line) + "\n")
			} else if isPlaying {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(genre.Color)).Render("  "+line) + "\n")
			} else {
				b.WriteString(normalItemStyle.Render(line) + "\n")
			}
		}

		b.WriteString("\n")
		b.WriteString(renderHints([]hint{
			{"↑/↓", "navigate"}, {"enter", "play"},
			{"s/a", "save to library"}, {"b/esc", "genres"},
			{"n", "now playing"}, {"h", "home"},
		}))
		return b.String()
	}

	// ── Genre picker ──────────────────────────────────────────────────────────
	b.WriteString(titleStyle.Render("✨  Recommended") + "\n")
	b.WriteString(subtitleStyle.Render("  Curated picks fetched live from YouTube — pick a genre") + "\n\n")

	genres := online.Genres

	// Render in a 2-column grid
	colWidth := 32
	for i, g := range genres {
		color := lipgloss.Color(g.Color)
		nameStr := g.Emoji + "  " + g.Name

		if i == a.recCursor {
			cell := lipgloss.NewStyle().
				Foreground(black).
				Background(color).
				Bold(true).
				Width(colWidth).
				Padding(0, 1).
				Render(nameStr)
			b.WriteString("  " + cell)
		} else {
			cell := lipgloss.NewStyle().
				Foreground(color).
				Width(colWidth).
				Padding(0, 1).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(color).
				Render(nameStr)
			b.WriteString("  " + cell)
		}

		// new line every 2 items for a 2-column layout
		if i%2 == 1 {
			b.WriteString("\n")
		}
	}
	if len(genres)%2 != 0 {
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderHints([]hint{
		{"↑/↓", "navigate"}, {"enter", "load picks"},
		{"n", "now playing"}, {"h", "back"},
	}))
	return b.String()
}
