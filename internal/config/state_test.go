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
