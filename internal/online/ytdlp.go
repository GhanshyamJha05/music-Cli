// Package online provides YouTube search and streaming via yt-dlp.
package online

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/GhanshyamJha05/music-Cli/internal/library"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// OnlineTrack is a normalised result from an online search.
type OnlineTrack struct {
	ID        string // source-specific id
	Title     string
	Artist    string
	Album     string
	Duration  int    // seconds
	StreamURL string // YouTube webpage URL
	Source    string // "ytdlp"
}

// TrackToSong converts an OnlineTrack to a library.Song so it can be played
// through the existing queue system.
func TrackToSong(t OnlineTrack) library.Song {
	return library.Song{
		ID:       "online_" + t.ID,
		Title:    t.Title,
		Artist:   t.Artist,
		Album:    t.Album,
		FilePath: t.StreamURL,
		Duration: t.Duration,
		Genre:    "Online / YouTube",
		AddedAt:  "-",
	}
}

// YtdlpAvailable checks if yt-dlp is on PATH.
func YtdlpAvailable() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

// FfmpegAvailable checks if ffmpeg is on PATH (required for audio extraction).
func FfmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// CheckDeps returns an error describing missing dependencies.
func CheckDeps() error {
	var missing []string
	if !YtdlpAvailable() {
		missing = append(missing, "yt-dlp (install from https://github.com/yt-dlp/yt-dlp)")
	}
	if !FfmpegAvailable() {
		missing = append(missing, "ffmpeg (required for audio extraction)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", "))
	}
	return nil
}

// maxRetries is the number of retry attempts for yt-dlp operations.
const maxRetries = 3

// retryDelay returns an exponential backoff duration for the given attempt.
func retryDelay(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

// YtdlpSearch searches YouTube via yt-dlp and returns up to maxResults tracks.
// It does NOT download audio - it only extracts metadata + URL.
// Retries up to maxRetries times with exponential backoff on transient errors.
func YtdlpSearch(query string, maxResults int) ([]OnlineTrack, error) {
	if err := CheckDeps(); err != nil {
		return nil, err
	}
	if maxResults <= 0 {
		maxResults = 10
	}

	searchArg := fmt.Sprintf("ytsearch%d:%s", maxResults, query)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay(attempt - 1))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		cmd := exec.CommandContext(ctx, "yt-dlp",
			"--dump-single-json",
			"--no-playlist",
			"--quiet",
			"--no-warnings",
			"--flat-playlist",
			searchArg,
		)
		out, err := cmd.Output()
		cancel()

		if err != nil {
			if ctx.Err() != nil {
				lastErr = fmt.Errorf("yt-dlp search timed out after 60s")
			} else {
				lastErr = fmt.Errorf("yt-dlp search error: %w", err)
			}
			continue
		}

		var playlist struct {
			Entries []struct {
				ID       string  `json:"id"`
				Title    string  `json:"title"`
				Uploader string  `json:"uploader"`
				Duration float64 `json:"duration"`
				URL      string  `json:"url"`
			} `json:"entries"`
		}

		if err := json.Unmarshal(out, &playlist); err != nil {
			return nil, fmt.Errorf("yt-dlp parse error: %w", err)
		}

		var tracks []OnlineTrack
		for _, e := range playlist.Entries {
			webURL := e.URL
			if webURL == "" {
				webURL = "https://www.youtube.com/watch?v=" + e.ID
			}
			tracks = append(tracks, OnlineTrack{
				ID:        e.ID,
				Title:     e.Title,
				Artist:    e.Uploader,
				Duration:  int(e.Duration),
				StreamURL: webURL,
				Source:    "ytdlp",
			})
		}
		return tracks, nil
	}
	return nil, fmt.Errorf("yt-dlp search failed after %d retries: %w", maxRetries, lastErr)
}

// YtdlpDownloadTemp downloads a YouTube URL to a temp MP3 file and returns
// the temp file path. The caller is responsible for deleting it after playback.
// Retries up to maxRetries times with exponential backoff on transient errors.
func YtdlpDownloadTemp(webpageURL string) (string, error) {
	if err := CheckDeps(); err != nil {
		return "", err
	}

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".music-Cli", "cache")
	_ = os.MkdirAll(cacheDir, 0755)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay(attempt - 1))
		}

		// Use a random name to prevent symlink attacks and parallel download collisions
		randBytes := make([]byte, 8)
		rand.Read(randBytes)
		outTemplate := filepath.Join(cacheDir, fmt.Sprintf("ytdlp_%s.%%(ext)s", hex.EncodeToString(randBytes)))

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		cmd := exec.CommandContext(ctx, "yt-dlp",
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "5",
			"--no-playlist",
			"--quiet",
			"--no-warnings",
			"-o", outTemplate,
			webpageURL,
		)
		err := cmd.Run()
		cancel()

		if err != nil {
			if ctx.Err() != nil {
				lastErr = fmt.Errorf("yt-dlp download timed out after 3 minutes")
			} else {
				lastErr = fmt.Errorf("yt-dlp download error: %w", err)
			}
			continue
		}

		// Find the file that was actually written
		matches, err := filepath.Glob(strings.Replace(outTemplate, "%(ext)s", "*", 1))
		if err != nil || len(matches) == 0 {
			mp3Path := strings.Replace(outTemplate, "%(ext)s", "mp3", 1)
			if _, err2 := os.Stat(mp3Path); err2 == nil {
				return mp3Path, nil
			}
			lastErr = fmt.Errorf("downloaded file not found")
			continue
		}
		return matches[0], nil
	}
	return "", fmt.Errorf("yt-dlp download failed after %d retries: %w", maxRetries, lastErr)
}

// CleanCache removes cached files older than 7 days.
// It respects the provided context for cancellation/timeout.
// Skips files that are locked or permission-denied (common on Windows).
func CleanCache(ctx context.Context) {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".music-Cli", "cache")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return // dir doesn't exist or can't be read — nothing to do
	}
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	for _, e := range entries {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue // can't stat — skip
		}
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(cacheDir, e.Name())
			// Retry once after a short delay (handles Windows file locks)
			if err := os.Remove(path); err != nil {
				time.Sleep(100 * time.Millisecond)
				os.Remove(path) // best-effort second attempt
			}
		}
	}
}
