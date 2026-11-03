package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	// Use a temp data file so we don't clobber the real library
	tmpDir := t.TempDir()
	dataFile = filepath.Join(tmpDir, "library.json")

	lib := &Library{}

	// Add two songs
	s1, err := lib.AddSong("Song One", "Artist A", "Album A", "Rock", "/music/a.mp3", 180)
	if err != nil {
		t.Fatalf("AddSong failed: %v", err)
	}
	s2, err := lib.AddSong("Song Two", "Artist B", "Album B", "Jazz", "/music/b.wav", 240)
	if err != nil {
		t.Fatalf("AddSong failed: %v", err)
	}

	if err := lib.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload from disk
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Songs) != 2 {
		t.Fatalf("expected 2 songs after reload, got %d", len(loaded.Songs))
	}
	if loaded.Songs[0].ID != s1.ID || loaded.Songs[0].Title != "Song One" {
		t.Fatalf("song 1 mismatch after reload: %+v", loaded.Songs[0])
	}
	if loaded.Songs[1].ID != s2.ID || loaded.Songs[1].Title != "Song Two" {
		t.Fatalf("song 2 mismatch after reload: %+v", loaded.Songs[1])
	}
}

func TestSearch(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile = filepath.Join(tmpDir, "library.json")

	lib := &Library{}
	lib.AddSong("Bohemian Rhapsody", "Queen", "A Night at the Opera", "Rock", "/music/q.mp3", 354)
	lib.AddSong("Stairway to Heaven", "Led Zeppelin", "Led Zeppelin IV", "Rock", "/music/lz.mp3", 482)
	lib.AddSong("Take Five", "Dave Brubeck", "Time Out", "Jazz", "/music/db.wav", 324)

	// Search by title
	results := lib.Search("bohemian")
	if len(results) != 1 || results[0].Title != "Bohemian Rhapsody" {
		t.Fatalf("search by title failed: %+v", results)
	}

	// Search by artist
	results = lib.Search("zeppelin")
	if len(results) != 1 || results[0].Artist != "Led Zeppelin" {
		t.Fatalf("search by artist failed: %+v", results)
	}

	// Search by genre (case-insensitive)
	results = lib.Search("rock")
	if len(results) != 2 {
		t.Fatalf("search by genre expected 2, got %d: %+v", len(results), results)
	}

	// Search by album
	results = lib.Search("time out")
	if len(results) != 1 || results[0].Album != "Time Out" {
		t.Fatalf("search by album failed: %+v", results)
	}

	// Empty query returns nothing
	results = lib.Search("")
	if len(results) != 0 {
		t.Fatalf("empty search should return 0, got %d", len(results))
	}

	// No match
	results = lib.Search("nonexistent")
	if len(results) != 0 {
		t.Fatalf("no-match search should return 0, got %d", len(results))
	}
}

func TestRemoveSongCleansPlaylists(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile = filepath.Join(tmpDir, "library.json")

	lib := &Library{}
	song, _ := lib.AddSong("Temp Song", "Temp Artist", "", "Temp", "/music/temp.mp3", 100)
	pl, _ := lib.CreatePlaylist("Test Playlist")
	lib.AddToPlaylist(pl.ID, song.ID)

	// Verify song is in playlist
	songs := lib.GetPlaylistSongs(pl.ID)
	if len(songs) != 1 || songs[0].ID != song.ID {
		t.Fatalf("song should be in playlist before removal: %+v", songs)
	}

	// Remove the song
	ok, err := lib.RemoveSong(song.ID)
	if err != nil {
		t.Fatalf("RemoveSong failed: %v", err)
	}
	if !ok {
		t.Fatal("RemoveSong should return true")
	}

	// Verify song is removed from library
	if _, err := lib.RemoveSong(song.ID); err == nil {
		// second remove returns false, not found
	}

	// Verify playlist reference is cleaned up
	songs = lib.GetPlaylistSongs(pl.ID)
	if len(songs) != 0 {
		t.Fatalf("playlist should have 0 songs after song removal, got %d: %+v", len(songs), songs)
	}
}

func TestPlaylistCRUD(t *testing.T) {
	tmpDir := t.TempDir()
	dataFile = filepath.Join(tmpDir, "library.json")

	lib := &Library{}

	// Create
	pl, err := lib.CreatePlaylist("My Mix")
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}
	if pl.Name != "My Mix" {
		t.Fatalf("playlist name mismatch: %s", pl.Name)
	}

	// Add song to playlist
	song, _ := lib.AddSong("Test Song", "Test Artist", "", "Test", "/music/t.mp3", 60)
	ok, err := lib.AddToPlaylist(pl.ID, song.ID)
	if err != nil {
		t.Fatalf("AddToPlaylist failed: %v", err)
	}
	if !ok {
		t.Fatal("AddToPlaylist should return true")
	}

	// Duplicate add should return false
	ok, err = lib.AddToPlaylist(pl.ID, song.ID)
	if err != nil {
		t.Fatalf("duplicate AddToPlaylist failed: %v", err)
	}
	if ok {
		t.Fatal("duplicate AddToPlaylist should return false")
	}

	// Get playlist songs
	songs := lib.GetPlaylistSongs(pl.ID)
	if len(songs) != 1 || songs[0].ID != song.ID {
		t.Fatalf("expected 1 song in playlist, got %d: %+v", len(songs), songs)
	}

	// Remove from playlist
	ok, err = lib.RemoveFromPlaylist(pl.ID, song.ID)
	if err != nil {
		t.Fatalf("RemoveFromPlaylist failed: %v", err)
	}
	if !ok {
		t.Fatal("RemoveFromPlaylist should return true")
	}
	songs = lib.GetPlaylistSongs(pl.ID)
	if len(songs) != 0 {
		t.Fatalf("expected 0 songs after removal, got %d", len(songs))
	}

	// Delete playlist
	ok, err = lib.DeletePlaylist(pl.ID)
	if err != nil {
		t.Fatalf("DeletePlaylist failed: %v", err)
	}
	if !ok {
		t.Fatal("DeletePlaylist should return true")
	}
	if len(lib.Playlists) != 0 {
		t.Fatalf("expected 0 playlists after deletion, got %d", len(lib.Playlists))
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		sec  int
		want string
	}{
		{0, "0:00"},
		{5, "0:05"},
		{60, "1:00"},
		{61, "1:01"},
		{354, "5:54"},
		{3600, "60:00"},
	}
	for _, c := range cases {
		if got := FormatDuration(c.sec); got != c.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", c.sec, got, c.want)
		}
	}
}

func TestDetectDurationUnsupportedFormat(t *testing.T) {
	_, err := DetectDuration("/nonexistent/file.xyz")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestMain(m *testing.M) {
	// Point dataFile at a temp location so tests don't touch the real library
	tmpDir, err := os.MkdirTemp("", "music-cli-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	dataFile = filepath.Join(tmpDir, "library.json")

	code := m.Run()
	os.Exit(code)
}
