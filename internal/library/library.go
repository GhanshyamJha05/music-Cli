package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
)

// idSeq is a monotonically increasing counter used to generate unique IDs
// even when AddSong/CreatePlaylist are called in rapid succession
// (time.Now().UnixNano() can return the same value within one nanosecond).
var idSeq int64

func nextID(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), atomic.AddInt64(&idSeq, 1))
}

// Song represents a single track
type Song struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	FilePath string `json:"file_path"`
	Duration int    `json:"duration"` // seconds
	Genre    string `json:"genre"`
	AddedAt  string `json:"added_at"`
}

// Playlist holds a named list of songs
type Playlist struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	SongIDs   []string `json:"song_ids"`
	CreatedAt string   `json:"created_at"`
}

// Library is the top-level data store
type Library struct {
	mu          sync.RWMutex         `json:"-"`
	Songs       []Song               `json:"songs"`
	Playlists   []Playlist           `json:"playlists"`
	searchIndex map[string][4]string // ID → [lowerTitle, lowerArtist, lowerAlbum, lowerGenre]
}

var dataFile string

func init() {
	home, err := os.UserHomeDir()
	var dataDir string
	if err != nil || home == "" {
		dataDir = ".music-Cli"
	} else {
		dataDir = filepath.Join(home, ".music-Cli")
	}
	_ = os.MkdirAll(dataDir, 0755)
	dataFile = filepath.Join(dataDir, "library.json")
}

// Load reads the library from disk, or returns an empty one
func Load() (*Library, error) {
	lib := &Library{}
	data, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		lib.buildSearchIndex()
		return lib, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, lib); err != nil {
		return nil, err
	}
	lib.buildSearchIndex()
	return lib, nil
}

// buildSearchIndex pre-computes lowercase fields for fast search
func (l *Library) buildSearchIndex() {
	l.searchIndex = make(map[string][4]string, len(l.Songs))
	for _, s := range l.Songs {
		l.searchIndex[s.ID] = [4]string{
			strings.ToLower(s.Title),
			strings.ToLower(s.Artist),
			strings.ToLower(s.Album),
			strings.ToLower(s.Genre),
		}
	}
}

// Save writes the library to disk
func (l *Library) Save() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.saveLocked()
}

// saveLocked writes the library to disk. Caller must hold at least a read lock.
func (l *Library) saveLocked() error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0644)
}

// AddSong adds a song and saves
func (l *Library) AddSong(title, artist, album, genre, filePath string, duration int) (Song, error) {
	song := Song{
		ID:       nextID("song"),
		Title:    title,
		Artist:   artist,
		Album:    album,
		Genre:    genre,
		FilePath: filePath,
		Duration: duration,
		AddedAt:  time.Now().Format("2006-01-02"),
	}
	l.mu.Lock()
	l.Songs = append(l.Songs, song)
	if l.searchIndex == nil {
		l.searchIndex = make(map[string][4]string)
	}
	l.searchIndex[song.ID] = [4]string{
		strings.ToLower(title),
		strings.ToLower(artist),
		strings.ToLower(album),
		strings.ToLower(genre),
	}
	err := l.saveLocked()
	l.mu.Unlock()
	return song, err
}

// RemoveSong removes a song by ID and cleans up any playlist references to it.
func (l *Library) RemoveSong(id string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, s := range l.Songs {
		if s.ID == id {
			l.Songs = append(l.Songs[:i], l.Songs[i+1:]...)
			delete(l.searchIndex, id)
			// Remove the song from every playlist that references it
			for pi := range l.Playlists {
				filtered := make([]string, 0, len(l.Playlists[pi].SongIDs))
				for _, sid := range l.Playlists[pi].SongIDs {
					if sid != id {
						filtered = append(filtered, sid)
					}
				}
				l.Playlists[pi].SongIDs = filtered
			}
			return true, l.saveLocked()
		}
	}
	return false, nil
}

// CreatePlaylist creates a new named playlist
func (l *Library) CreatePlaylist(name string) (Playlist, error) {
	pl := Playlist{
		ID:        nextID("pl"),
		Name:      name,
		SongIDs:   []string{},
		CreatedAt: time.Now().Format("2006-01-02"),
	}
	l.mu.Lock()
	l.Playlists = append(l.Playlists, pl)
	err := l.saveLocked()
	l.mu.Unlock()
	return pl, err
}

// DeletePlaylist removes a playlist by ID
func (l *Library) DeletePlaylist(id string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, p := range l.Playlists {
		if p.ID == id {
			l.Playlists = append(l.Playlists[:i], l.Playlists[i+1:]...)
			return true, l.saveLocked()
		}
	}
	return false, nil
}

// AddToPlaylist adds a song to a playlist
func (l *Library) AddToPlaylist(playlistID, songID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, p := range l.Playlists {
		if p.ID == playlistID {
			for _, sid := range p.SongIDs {
				if sid == songID {
					return false, nil // already in
				}
			}
			l.Playlists[i].SongIDs = append(l.Playlists[i].SongIDs, songID)
			return true, l.saveLocked()
		}
	}
	return false, nil
}

// RemoveFromPlaylist removes a song from a playlist
func (l *Library) RemoveFromPlaylist(playlistID, songID string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, p := range l.Playlists {
		if p.ID == playlistID {
			for j, sid := range p.SongIDs {
				if sid == songID {
					l.Playlists[i].SongIDs = append(p.SongIDs[:j], p.SongIDs[j+1:]...)
					return true, l.saveLocked()
				}
			}
		}
	}
	return false, nil
}

// SongExists checks if a song already exists in the library by file path or (title + artist)
func (l *Library) SongExists(filePath, title, artist string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	normFP := strings.TrimSpace(strings.ToLower(filePath))
	normTitle := strings.TrimSpace(strings.ToLower(title))
	normArtist := strings.TrimSpace(strings.ToLower(artist))

	for _, s := range l.Songs {
		// Exact file path or URL match
		if normFP != "" && strings.TrimSpace(strings.ToLower(s.FilePath)) == normFP {
			return true
		}
		// Exact title & artist match
		if normTitle != "" && strings.TrimSpace(strings.ToLower(s.Title)) == normTitle {
			if normArtist == "" || strings.TrimSpace(strings.ToLower(s.Artist)) == normArtist {
				return true
			}
		}
	}
	return false
}

// MoveSong shifts a song's position in the library slice (for reordering).
func (l *Library) MoveSong(fromIndex, toIndex int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if fromIndex < 0 || fromIndex >= len(l.Songs) || toIndex < 0 || toIndex >= len(l.Songs) || fromIndex == toIndex {
		return false
	}
	song := l.Songs[fromIndex]
	l.Songs = append(l.Songs[:fromIndex], l.Songs[fromIndex+1:]...)
	l.Songs = append(l.Songs[:toIndex], append([]Song{song}, l.Songs[toIndex:]...)...)
	_ = l.saveLocked()
	return true
}

// MovePlaylistSong shifts a song's position inside a playlist.
func (l *Library) MovePlaylistSong(playlistID string, fromIndex, toIndex int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	for pi, pl := range l.Playlists {
		if pl.ID == playlistID {
			if fromIndex < 0 || fromIndex >= len(pl.SongIDs) || toIndex < 0 || toIndex >= len(pl.SongIDs) || fromIndex == toIndex {
				return false
			}
			sid := pl.SongIDs[fromIndex]
			l.Playlists[pi].SongIDs = append(pl.SongIDs[:fromIndex], pl.SongIDs[fromIndex+1:]...)
			l.Playlists[pi].SongIDs = append(l.Playlists[pi].SongIDs[:toIndex], append([]string{sid}, l.Playlists[pi].SongIDs[toIndex:]...)...)
			_ = l.saveLocked()
			return true
		}
	}
	return false
}

// GetPlaylistSongs returns the songs belonging to a playlist
func (l *Library) GetPlaylistSongs(playlistID string) []Song {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, p := range l.Playlists {
		if p.ID == playlistID {
			if len(p.SongIDs) == 0 {
				return nil
			}
			songMap := make(map[string]Song, len(l.Songs))
			for _, s := range l.Songs {
				songMap[s.ID] = s
			}
			result := make([]Song, 0, len(p.SongIDs))
			for _, sid := range p.SongIDs {
				if s, ok := songMap[sid]; ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

// Search returns songs matching the query (title, artist, album, genre).
// An empty query returns no results.
func (l *Library) Search(query string) []Song {
	l.mu.RLock()
	defer l.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	results := make([]Song, 0, 16)
	for _, s := range l.Songs {
		fields, ok := l.searchIndex[s.ID]
		if !ok {
			continue
		}
		if strings.Contains(fields[0], q) ||
			strings.Contains(fields[1], q) ||
			strings.Contains(fields[2], q) ||
			strings.Contains(fields[3], q) {
			results = append(results, s)
		}
	}
	return results
}

// FormatDuration converts seconds to mm:ss
func FormatDuration(sec int) string {
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// DetectDuration reads an audio file and returns its duration in seconds.
// Supports MP3, WAV, and FLAC formats.
func DetectDuration(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	default:
		return 0, fmt.Errorf("unsupported format: %s", ext)
	}

	if err != nil {
		return 0, fmt.Errorf("decode error: %w", err)
	}
	defer streamer.Close()

	// Duration = length / sample rate
	duration := format.SampleRate.D(streamer.Len())
	return int(duration.Seconds()), nil
}
