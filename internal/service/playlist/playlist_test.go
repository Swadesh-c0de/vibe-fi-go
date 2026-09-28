package playlist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlaylistManager(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	mgr := NewPlaylistManager()

	if err := mgr.CreatePlaylist("Synthwave"); err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	song := PlaylistSong{
		Title:    "Resonance",
		URL:      "https://example.com/resonance.mp3",
		Duration: "03:32",
	}
	if err := mgr.AddSongToPlaylist("Synthwave", song); err != nil {
		t.Fatalf("AddSongToPlaylist failed: %v", err)
	}

	songs := mgr.GetPlaylistSongs("Synthwave")
	if len(songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(songs))
	}
	if songs[0].Title != "Resonance" {
		t.Errorf("expected title Resonance, got %q", songs[0].Title)
	}

	// Test M3U export
	m3uPath := filepath.Join(tempDir, "test.m3u")
	if err := mgr.ExportToM3U("Synthwave", m3uPath); err != nil {
		t.Fatalf("ExportToM3U failed: %v", err)
	}
	content, err := os.ReadFile(m3uPath)
	if err != nil {
		t.Fatalf("failed to read m3u: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("m3u file is empty")
	}
}
