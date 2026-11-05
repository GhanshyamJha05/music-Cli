package player

import (
	"math"
	"testing"
)

func approx(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestVolumeClamping(t *testing.T) {
	p := Get()

	// Volume above 1.0 should clamp to 1.0
	p.SetVolume(2.0)
	if v := p.GetVolume(); v != 1.0 {
		t.Fatalf("SetVolume(2.0) clamped to %v, want 1.0", v)
	}

	// Volume below 0.0 should clamp to 0.0
	p.SetVolume(-1.0)
	if v := p.GetVolume(); v != 0.0 {
		t.Fatalf("SetVolume(-1.0) clamped to %v, want 0.0", v)
	}

	// Normal volume should be preserved
	p.SetVolume(0.5)
	if v := p.GetVolume(); v != 0.5 {
		t.Fatalf("SetVolume(0.5) should be 0.5, got %v", v)
	}
}

func TestVolumeUpDown(t *testing.T) {
	p := Get()
	p.SetVolume(0.7)

	p.VolumeUp()
	if v := p.GetVolume(); !approx(v, 0.75) {
		t.Fatalf("VolumeUp from 0.7 should be 0.75, got %v", v)
	}

	p.VolumeDown()
	p.VolumeDown()
	if v := p.GetVolume(); !approx(v, 0.65) {
		t.Fatalf("VolumeDown twice from 0.75 should be 0.65, got %v", v)
	}
}

func TestInitialStateStopped(t *testing.T) {
	p := Get()
	if p.GetState() != StateStopped {
		t.Fatalf("initial state should be Stopped, got %v", p.GetState())
	}
	if p.IsPlaying() {
		t.Fatal("IsPlaying should be false initially")
	}
	if p.IsPaused() {
		t.Fatal("IsPaused should be false initially")
	}
}

func TestSeekOnStoppedPlayerIsNoop(t *testing.T) {
	p := Get()
	// Seek should not panic on a stopped player
	p.Seek(10 * 1000000000)  // 10 seconds forward
	p.Seek(-10 * 1000000000) // 10 seconds backward
	if p.GetState() != StateStopped {
		t.Fatalf("state should remain Stopped after seek, got %v", p.GetState())
	}
}
