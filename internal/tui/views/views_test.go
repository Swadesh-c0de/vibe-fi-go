package views

import (
	"strings"
	"testing"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/tui/visualizer"
)

func TestAllViewsRendering(t *testing.T) {
	mock := player.NewMockPlayer()
	_ = mock.Load("song.mp3", "replace")
	_ = mock.Play()

	th := theme.GetTheme("Midnight")
	styles := theme.MakeStyles(th)
	viz := visualizer.NewVisualizer()

	w := 80
	h := 20

	// 1. Playback View
	lyricsData := lyrics.LyricsData{
		HasSynced: true,
		SyncedLyrics: []lyrics.LyricLine{
			{Timestamp: 0.0, Text: "Intro line"},
			{Timestamp: 5.0, Text: "Second line of the song"},
			{Timestamp: 10.0, Text: "Third line continuing"},
		},
	}
	pb, _ := RenderPlaybackView(w, h, mock, viz, lyricsData, 0, true, styles)
	lines := strings.Split(pb, "\n")
	if len(lines) != h {
		t.Errorf("PlaybackView: expected %d lines, got %d", h, len(lines))
	}

	// 2. Library View
	libItems := []library.LibraryItem{
		{Name: "MusicFolder", Path: "/MusicFolder", IsDirectory: true},
		{Name: "Track1.mp3", Path: "/Track1.mp3", Duration: "03:45", IsDirectory: false},
	}
	lv := RenderLibraryView(w, h, "/home/music", libItems, 0, 0, styles)
	if len(strings.Split(lv, "\n")) != h {
		t.Errorf("LibraryView: expected %d lines", h)
	}

	// 3. Search Views
	si := RenderSearchInput(w, h, "lofi beats", styles)
	if len(strings.Split(si, "\n")) != h {
		t.Errorf("SearchInput: expected %d lines", h)
	}

	searchResults := []search.SearchResult{
		{Title: "Chill Lofi Beat", URL: "https://youtube.com/1", Duration: "02:30", Uploader: "Artist"},
	}
	sr := RenderSearchResults(w, h, searchResults, 0, 0, styles)
	if len(strings.Split(sr, "\n")) != h {
		t.Errorf("SearchResults: expected %d lines", h)
	}

	// 4. Playlist Views
	playlists := []playlist.Playlist{
		{Name: "Chill", Path: "Chill.txt", SongCount: 5},
	}
	previewSongs := []playlist.PlaylistSong{
		{Title: "Song 1", URL: "url1", Duration: "03:00"},
	}
	pbView := RenderPlaylistsBrowser(w, h, playlists, previewSongs, 0, 0, styles)
	if len(strings.Split(pbView, "\n")) != h {
		t.Errorf("PlaylistsBrowser: expected %d lines", h)
	}

	// 5. Intro View
	intro := RenderIntroView(w, h, styles)
	if len(strings.Split(intro, "\n")) != h {
		t.Errorf("IntroView: expected %d lines", h)
	}
}
