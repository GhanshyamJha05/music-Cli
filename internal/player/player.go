package player

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// State represents the playback state
type State int

const (
	StateStopped State = iota
	StatePlaying
	StatePaused
)

// playSession groups the done channel with a sync.Once to guarantee
// the channel is closed exactly once, even if stopLocked() and the
// speaker callback fire concurrently.
type playSession struct {
	done chan struct{}
	once sync.Once
}

func (s *playSession) close() {
	s.once.Do(func() { close(s.done) })
}

// volStreamer wraps a beep.Streamer and applies volume scaling.
type volStreamer struct {
	streamer beep.Streamer
	volume   float64 // 0.0 to 1.0
}

func (v *volStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.streamer.Stream(samples)
	if v.volume < 1.0 {
		for i := range samples[:n] {
			samples[i][0] *= v.volume
			samples[i][1] *= v.volume
		}
	}
	return n, ok
}

func (v *volStreamer) Err() error {
	return v.streamer.Err()
}

func (v *volStreamer) SetVolume(vol float64) {
	v.volume = vol
}

// Player manages audio playback
type Player struct {
	mu             sync.Mutex
	state          State
	currentFile    string
	ctrl           *beep.Ctrl
	volStr         *volStreamer
	streamer       beep.StreamSeekCloser
	session        *playSession
	started        time.Time
	pausedAt       time.Duration
	OnFinish       func()
	currentRate    beep.SampleRate
	volume         float64 // 0.0 (silent) to 1.0 (full)
	tempFiles      map[string]struct{} // tracks active temp files
}

var globalPlayer *Player
var once sync.Once
var speakerOnce sync.Once

// Get returns the singleton player
func Get() *Player {
	once.Do(func() {
		globalPlayer = &Player{
			tempFiles: make(map[string]struct{}),
		}
	})
	return globalPlayer
}

// Play starts playing a file
func (p *Player) Play(filePath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playLocked(filePath)
}

func (p *Player) playLocked(filePath string) error {
	// Stop any current playback
	p.stopLocked()

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}

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
		f.Close()
		return fmt.Errorf("unsupported format: %s", ext)
	}

	if err != nil {
		f.Close()
		return fmt.Errorf("decode error: %w", err)
	}

	// speaker.Init must be called exactly once. Initialize with the first
	// track's sample rate; do not re-initialize when switching tracks.
	speakerOnce.Do(func() {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		p.currentRate = format.SampleRate
	})

	p.ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
	p.streamer = streamer
	p.currentFile = filePath
	p.state = StatePlaying
	p.started = time.Now()
	p.pausedAt = 0

	// Initialize volume (default 0.7)
	if p.volume == 0 {
		p.volume = 0.7
	}
	p.volStr = &volStreamer{
		streamer: p.ctrl,
		volume:   p.volume,
	}

	sess := &playSession{done: make(chan struct{})}
	p.session = sess

	onFinish := p.OnFinish

	speaker.Play(beep.Seq(p.volStr, beep.Callback(func() {
		sess.close()
		if onFinish != nil {
			onFinish()
		}
	})))

	return nil
}

// Pause toggles pause/resume
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	if p.ctrl.Paused {
		p.pausedAt = p.pausedAt + time.Since(p.started)
		p.state = StatePaused
	} else {
		p.started = time.Now()
		p.state = StatePlaying
	}
	speaker.Unlock()
}

// Stop halts playback
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

func (p *Player) stopLocked() {
	if p.session != nil {
		p.session.close()
		p.session = nil
	}
	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
		p.ctrl = nil
	}
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	p.state = StateStopped
	p.currentFile = ""
	p.pausedAt = 0
}

// State returns current playback state
func (p *Player) GetState() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// CurrentFile returns the currently playing file path
func (p *Player) CurrentFile() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentFile
}

// Elapsed returns how long the current track has been playing
func (p *Player) Elapsed() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state == StatePlaying {
		return p.pausedAt + time.Since(p.started)
	}
	return p.pausedAt
}

// IsPlaying returns true if audio is actively playing
func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state == StatePlaying
}

// IsPaused returns true if paused
func (p *Player) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state == StatePaused
}

// SetVolume sets the playback volume (0.0 to 1.0).
func (p *Player) SetVolume(vol float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	p.volume = vol
	if p.volStr != nil {
		speaker.Lock()
		p.volStr.SetVolume(vol)
		speaker.Unlock()
	}
}

// VolumeUp increases volume by 5%.
func (p *Player) VolumeUp() {
	p.SetVolume(p.volume + 0.05)
}

// VolumeDown decreases volume by 5%.
func (p *Player) VolumeDown() {
	p.SetVolume(p.volume - 0.05)
}

// GetVolume returns current volume (0.0 to 1.0).
func (p *Player) GetVolume() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

// Seek jumps forward or backward by the given duration.
// Positive duration = forward, negative = backward.
func (p *Player) Seek(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.streamer == nil || p.state == StateStopped {
		return
	}

	// Calculate current position in samples
	rate := p.currentRate
	if rate == 0 {
		return
	}

	currentPos := p.Elapsed()
	newPos := currentPos + d
	if newPos < 0 {
		newPos = 0
	}

	samplePos := rate.N(newPos)
	speaker.Lock()
	err := p.streamer.Seek(samplePos)
	if err == nil {
		// Update pausedAt so Elapsed() returns the correct position
		p.pausedAt = newPos
		if p.state == StatePlaying {
			p.started = time.Now()
		}
	}
	speaker.Unlock()
}

// SeekTo jumps to an absolute position in the track.
func (p *Player) SeekTo(pos time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.streamer == nil || p.state == StateStopped {
		return
	}
	rate := p.currentRate
	if rate == 0 {
		return
	}
	if pos < 0 {
		pos = 0
	}
	samplePos := rate.N(pos)
	speaker.Lock()
	err := p.streamer.Seek(samplePos)
	if err == nil {
		p.pausedAt = pos
		if p.state == StatePlaying {
			p.started = time.Now()
		}
	}
	speaker.Unlock()
}

// PlayTempFile plays a local file and schedules its deletion after playback.
// Each download uses a unique random name, so concurrent playbacks are safe.
// Files are cleaned up by the background cache cleaner, not eagerly deleted.
func (p *Player) PlayTempFile(path string) error {
	p.mu.Lock()
	if err := p.playLocked(path); err != nil {
		p.mu.Unlock()
		os.Remove(path)
		return err
	}
	p.tempFiles[path] = struct{}{}
	p.mu.Unlock()

	// Clean up the temp file after playback finishes (best-effort, non-blocking)
	go func() {
		// Wait for playback to end
		p.mu.Lock()
		sess := p.session
		p.mu.Unlock()
		if sess != nil {
			<-sess.done
		}
		// Small delay to ensure file handles are released
		time.Sleep(200 * time.Millisecond)
		os.Remove(path)
		p.mu.Lock()
		delete(p.tempFiles, path)
		p.mu.Unlock()
	}()
	return nil
}
