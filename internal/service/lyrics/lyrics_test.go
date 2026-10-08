package lyrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSyncedLyrics(t *testing.T) {
	rawLRC := `[ti:Test Song]
[ar:Test Artist]
[00:01.20]First line of lyrics
[00:05.50]Second line of lyrics
[01:10.00]Third line after one minute
`
	lines := ParseSyncedLyrics(rawLRC)
	if len(lines) != 3 {
		t.Fatalf("expected 3 synced lines, got %d", len(lines))
	}

	if lines[0].Text != "First line of lyrics" || lines[0].Timestamp < 1.19 || lines[0].Timestamp > 1.21 {
		t.Errorf("line 0 mismatch: %+v", lines[0])
	}
	if lines[1].Text != "Second line of lyrics" || lines[1].Timestamp < 5.49 || lines[1].Timestamp > 5.51 {
		t.Errorf("line 1 mismatch: %+v", lines[1])
	}
	if lines[2].Text != "Third line after one minute" || lines[2].Timestamp < 69.9 || lines[2].Timestamp > 70.1 {
		t.Errorf("line 2 mismatch: %+v", lines[2])
	}
}

func TestLoadLocalLRC(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "MySong.mp3")
	lrcPath := filepath.Join(tmpDir, "MySong.lrc")

	// Create dummy audio and .lrc files
	_ = os.WriteFile(audioPath, []byte("fake mp3 data"), 0644)
	lrcContent := "[00:02.00]Hello world\n[00:04.50]Goodbye world\n"
	_ = os.WriteFile(lrcPath, []byte(lrcContent), 0644)

	mgr := NewLyricsManager()
	data, ok := mgr.LoadLocalLRC(audioPath, "MySong")
	if !ok {
		t.Fatalf("expected LoadLocalLRC to succeed for companion .lrc")
	}
	if !data.HasSynced {
		t.Errorf("expected HasSynced=true")
	}
	if len(data.SyncedLyrics) != 2 {
		t.Errorf("expected 2 synced lines, got %d", len(data.SyncedLyrics))
	}
	if data.SyncedLyrics[0].Text != "Hello world" {
		t.Errorf("unexpected first line: %s", data.SyncedLyrics[0].Text)
	}

	// Test non-existent file
	_, notOk := mgr.LoadLocalLRC(filepath.Join(tmpDir, "OtherSong.mp3"), "OtherSong")
	if notOk {
		t.Errorf("expected LoadLocalLRC to fail for missing file")
	}

	// Test URL input
	_, urlOk := mgr.LoadLocalLRC("https://youtube.com/watch?v=12345", "")
	if urlOk {
		t.Errorf("expected LoadLocalLRC to reject URLs")
	}
}

func TestFetchLyricsPrioritizesSynced(t *testing.T) {
	mgr := NewLyricsManager()
	data, err := mgr.FetchLyrics("Ruth B.", "Dandelions", 235)
	if err != nil {
		t.Skipf("Network unavailable: %v", err)
	}
	if !data.HasSynced {
		t.Errorf("expected HasSynced=true for Ruth B. Dandelions after prioritizing search over plain exact")
	}
	if len(data.SyncedLyrics) == 0 {
		t.Errorf("expected non-empty SyncedLyrics")
	}
}

