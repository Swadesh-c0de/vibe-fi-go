package views

import (
	"strings"
	"testing"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/tui/visualizer"
	"vibe-fi/internal/utils/stringutil"
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
	pb, _, _ := RenderPlaybackView(w, h, mock, viz, lyricsData, 0, true, styles)
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

func TestLyricsAutoScrollBoundary(t *testing.T) {
	themeObj := theme.Themes[0]
	styles := theme.MakeStyles(themeObj)
	mock := player.NewMockPlayer()
	_ = mock.Load("song.mp3", "replace")
	_ = mock.Seek(50.0)

	// Create 20 lines of synced lyrics
	var synLyrics []lyrics.LyricLine
	for i := 0; i < 20; i++ {
		synLyrics = append(synLyrics, lyrics.LyricLine{
			Timestamp: float64(i * 5),
			Text:      "छोड़ के न चल पड़य तू line",
		})
	}
	lyricsData := lyrics.LyricsData{
		HasSynced:    true,
		SyncedLyrics: synLyrics,
	}

	// activeIdx at pos 50.0 will be 10 (10*5 = 50.0).
	// With textW=60, textH=8, offset=0, autoScroll=false:
	// activeIdx (10) is >= offset + textH - 1 (7), so it reached the end of visible lines!
	// It should trigger auto-scroll and return newAutoScroll=true, with newOffset centering activeIdx.
	out, newOffset, newAutoScroll := renderLyricsBody(60, 8, mock, lyricsData, 0, false, styles)
	if !newAutoScroll {
		t.Errorf("expected autoScroll to re-engage when active lyric reaches end of visible lines")
	}
	if newOffset <= 0 {
		t.Errorf("expected newOffset > 0, got %d", newOffset)
	}
	if len(out) != 8 {
		t.Errorf("expected 8 lines, got %d", len(out))
	}
}

func TestDevanagariLyricsBoxAlignment(t *testing.T) {
	themeObj := theme.Themes[0]
	styles := theme.MakeStyles(themeObj)
	mock := player.NewMockPlayer()
	_ = mock.Load("song.mp3", "replace")

	hindiLines := []string{
		"छोड़ के न चल पड़य तू",
		"छोड़ के न चल पड़य तू",
		"कण कड़े रात खवेगी",
		"तन्ने मेरी याद, आवैगी",
		"के तन्ने मेरी याद, आवैगी?",
		"तन्ने मेरी याद आवैगी",
		"य ते मन्ने मर खंव रै",
		"स्यूं मन्ने, छोड़ के गय?",
	}

	lyricsData := lyrics.LyricsData{
		PlainLyrics: strings.Join(hindiLines, "\n"),
	}

	boxW := 80
	boxH := 10
	outLines, _, _ := renderLyricsBody(boxW-2, boxH-2, mock, lyricsData, 0, false, styles)
	box := components.RenderBoxWithTitle("LYRICS", outLines, boxW, boxH, styles)

	rows := strings.Split(box, "\n")
	if len(rows) != boxH {
		t.Fatalf("expected %d rows, got %d", boxH, len(rows))
	}

	for i, row := range rows {
		w := stringutil.Width(row)
		if w != boxW {
			t.Errorf("Row %d width = %d, expected %d. Content: %q", i, w, boxW, row)
		}
	}
}
