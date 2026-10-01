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

func TestParsePlaylistLineWithPipes(t *testing.T) {
	line := "BEAUZ & Heleen - Alone | Trap | NCS - Copyright Free Music|https://www.youtube.com/watch?v=A2xYanbip1w|2:38"
	song, ok := ParsePlaylistLine(line)
	if !ok {
		t.Fatalf("failed to parse valid line with pipes")
	}

	expectedTitle := "BEAUZ & Heleen - Alone | Trap | NCS - Copyright Free Music"
	expectedURL := "https://www.youtube.com/watch?v=A2xYanbip1w"
	expectedDur := "2:38"

	if song.Title != expectedTitle {
		t.Errorf("expected title %q, got %q", expectedTitle, song.Title)
	}
	if song.URL != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, song.URL)
	}
	if song.Duration != expectedDur {
		t.Errorf("expected duration %q, got %q", expectedDur, song.Duration)
	}
}

func TestDuplicateURLAvoidance(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	mgr := NewPlaylistManager()
	_ = mgr.CreatePlaylist("Dupes")

	song := PlaylistSong{
		Title:    "Track 1",
		URL:      "https://example.com/same.mp3",
		Duration: "2:00",
	}

	_ = mgr.AddSongToPlaylist("Dupes", song)
	_ = mgr.AddSongToPlaylist("Dupes", song) // Should be ignored

	songs := mgr.GetPlaylistSongs("Dupes")
	if len(songs) != 1 {
		t.Errorf("expected 1 song after duplicate add, got %d", len(songs))
	}
}

func TestJSONPlaylistFallback(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	pDir := filepath.Join(tempDir, ".vibe-fi", "playlists")
	_ = os.MkdirAll(pDir, 0755)

	jsonContent := `{
  "name": "OnlyJSON",
  "songs": [
    {
      "title": "Acoustic Morning",
      "url": "https://example.com/acoustic.mp3",
      "duration": "3:15"
    }
  ]
}`
	_ = os.WriteFile(filepath.Join(pDir, "OnlyJSON.json"), []byte(jsonContent), 0644)

	mgr := NewPlaylistManager()
	pls := mgr.ListPlaylists()
	found := false
	for _, pl := range pls {
		if pl.Name == "OnlyJSON" && pl.SongCount == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find OnlyJSON in ListPlaylists")
	}

	songs := mgr.GetPlaylistSongs("OnlyJSON")
	if len(songs) != 1 {
		t.Fatalf("expected 1 song from JSON playlist, got %d", len(songs))
	}
	if songs[0].Title != "Acoustic Morning" {
		t.Errorf("expected title Acoustic Morning, got %q", songs[0].Title)
	}
	if songs[0].URL != "https://example.com/acoustic.mp3" {
		t.Errorf("expected URL https://example.com/acoustic.mp3, got %q", songs[0].URL)
	}
}
