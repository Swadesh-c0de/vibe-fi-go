package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateLoadSave(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	state := SessionState{
		Path:       "https://example.com/audio.mp3",
		Title:      "Test Song",
		Position:   45.5,
		Volume:     85,
		Playlist:   "Favorites",
		Index:      2,
		Theme:      "Matrix",
		Visualizer: 1,
		Autoplay:   true,
	}

	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	stateFile := filepath.Join(tempDir, ".vibe-fi", "state.ini")
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("state.ini not created: %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.Title != state.Title {
		t.Errorf("expected Title %q, got %q", state.Title, loaded.Title)
	}
	if loaded.Theme != state.Theme {
		t.Errorf("expected Theme %q, got %q", state.Theme, loaded.Theme)
	}
	if loaded.Volume != state.Volume {
		t.Errorf("expected Volume %d, got %d", state.Volume, loaded.Volume)
	}
	if loaded.Visualizer != state.Visualizer {
		t.Errorf("expected Visualizer %d, got %d", state.Visualizer, loaded.Visualizer)
	}
	if loaded.Autoplay != state.Autoplay {
		t.Errorf("expected Autoplay %v, got %v", state.Autoplay, loaded.Autoplay)
	}
}

func TestDefaultState(t *testing.T) {
	def := DefaultState()
	if !def.Autoplay {
		t.Errorf("expected default Autoplay to be true, got %v", def.Autoplay)
	}
	if def.Volume != 100 {
		t.Errorf("expected default Volume to be 100, got %d", def.Volume)
	}
}

func TestStateLoadSaveAutoplayFalse(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// When no state.ini exists, LoadState returns DefaultState (Autoplay: true, Volume: 100)
	initial, err := LoadState()
	if err == nil { // no error or file not found fallback
		if !initial.Autoplay {
			t.Errorf("expected initial Autoplay true, got %v", initial.Autoplay)
		}
		if initial.Volume != 100 {
			t.Errorf("expected initial Volume 100, got %d", initial.Volume)
		}
	}

	// Now save state with Autoplay: false, Volume: 65
	st := SessionState{
		Volume:   65,
		Autoplay: false,
	}
	if err := SaveState(st); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if loaded.Autoplay {
		t.Errorf("expected Autoplay false after saving false, got %v", loaded.Autoplay)
	}
	if loaded.Volume != 65 {
		t.Errorf("expected Volume 65 after saving 65, got %d", loaded.Volume)
	}
}
