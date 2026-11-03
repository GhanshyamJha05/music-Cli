package online

import (
	"testing"
)

func TestYtdlpAvailableSkip(t *testing.T) {
	// This test documents the dependency-check behavior.
	// It passes regardless of whether yt-dlp is installed; if it's not,
	// online features are correctly reported as unavailable.
	available := YtdlpAvailable()
	t.Logf("yt-dlp available: %v", available)

	if !available {
		t.Skip("yt-dlp not installed — skipping yt-dlp dependent tests")
	}
}

func TestFfmpegAvailableSkip(t *testing.T) {
	available := FfmpegAvailable()
	t.Logf("ffmpeg available: %v", available)

	if !available {
		t.Skip("ffmpeg not installed — skipping ffmpeg dependent tests")
	}
}

func TestCheckDepsMissing(t *testing.T) {
	// CheckDeps returns an error describing missing dependencies.
	// We can't control the environment, so just verify it returns either
	// nil or an error containing the dependency names.
	err := CheckDeps()
	if err != nil {
		t.Logf("CheckDeps reported missing: %v", err)
		// The error should mention yt-dlp and/or ffmpeg
		msg := err.Error()
		if !contains(msg, "yt-dlp") && !contains(msg, "ffmpeg") {
			t.Fatalf("CheckDeps error should mention yt-dlp or ffmpeg, got: %s", msg)
		}
	}
}

func TestTrackToSong(t *testing.T) {
	track := OnlineTrack{
		ID:        "abc123",
		Title:     "Test Title",
		Artist:    "Test Artist",
		Album:     "Test Album",
		Duration:  180,
		StreamURL: "https://www.youtube.com/watch?v=abc123",
		Source:    "ytdlp",
	}

	song := TrackToSong(track)
	if song.ID != "online_abc123" {
		t.Fatalf("expected ID online_abc123, got %s", song.ID)
	}
	if song.Title != "Test Title" {
		t.Fatalf("expected title Test Title, got %s", song.Title)
	}
	if song.Artist != "Test Artist" {
		t.Fatalf("expected artist Test Artist, got %s", song.Artist)
	}
	if song.Album != "Test Album" {
		t.Fatalf("expected album Test Album, got %s", song.Album)
	}
	if song.Duration != 180 {
		t.Fatalf("expected duration 180, got %d", song.Duration)
	}
	if song.FilePath != "https://www.youtube.com/watch?v=abc123" {
		t.Fatalf("expected FilePath to be the stream URL, got %s", song.FilePath)
	}
	if song.Genre != "Online / YouTube" {
		t.Fatalf("expected genre Online / YouTube, got %s", song.Genre)
	}
}

func TestGenresNonEmpty(t *testing.T) {
	if len(Genres) == 0 {
		t.Fatal("Genres list should not be empty")
	}
	for _, g := range Genres {
		if g.Name == "" {
			t.Fatal("genre should have a non-empty name")
		}
		if g.Query == "" {
			t.Fatalf("genre %s should have a non-empty query", g.Name)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
