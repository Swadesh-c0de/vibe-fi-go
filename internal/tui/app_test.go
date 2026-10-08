package tui

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"vibe-fi/internal/eventbus"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
)

func TestAppModelEventBus(t *testing.T) {
	mock := player.NewMockPlayer()
	app := NewAppModel(mock)

	if app.EventBus == nil {
		t.Fatalf("expected EventBus to be initialized on AppModel")
	}

	var startedEvent *eventbus.TrackStartedEvent
	app.EventBus.Subscribe(eventbus.EventTrackStarted, func(e interface{}) {
		if evt, ok := e.(eventbus.TrackStartedEvent); ok {
			startedEvent = &evt
		}
	})

	var playbackState eventbus.PlaybackState
	app.EventBus.Subscribe(eventbus.EventPlaybackStateChanged, func(e interface{}) {
		if evt, ok := e.(eventbus.PlaybackStateChangedEvent); ok {
			playbackState = evt.State
		}
	})

	_ = app.StartTrackPlayback("Chill Beats", "/path/to/chill.mp3", "02:30", "Artist A")

	if startedEvent == nil {
		t.Fatalf("expected EventTrackStarted to be published")
	}
	if startedEvent.Title != "Chill Beats" {
		t.Errorf("expected Title 'Chill Beats', got %q", startedEvent.Title)
	}
	if startedEvent.URL != "/path/to/chill.mp3" {
		t.Errorf("expected URL '/path/to/chill.mp3', got %q", startedEvent.URL)
	}
	if playbackState != eventbus.StatePlaying {
		t.Errorf("expected StatePlaying, got %v", playbackState)
	}
}

func TestAppModelViewRendering(t *testing.T) {
	mock := player.NewMockPlayer()
	app := NewAppModel(mock)
	app.Width = 80
	app.Height = 24

	// 1. Intro View
	app.SetMode(components.ViewModeIntro)
	v := app.View()
	if v == "" {
		t.Errorf("expected non-empty View in Intro mode")
	}

	// 2. Playback View
	app.SetMode(components.ViewModePlayback)
	v = app.View()
	if v == "" {
		t.Errorf("expected non-empty View in Playback mode")
	}

	// 3. Help Modal
	app.ShowHelpModal = true
	v = app.View()
	if v == "" {
		t.Errorf("expected non-empty View with Help Modal")
	}
}

func TestIntroInteraction(t *testing.T) {
	mock := player.NewMockPlayer()
	app := NewAppModel(mock)
	app.SetMode(components.ViewModeIntro)

	if app.SelectionIndex != 0 {
		t.Errorf("expected initial SelectionIndex 0, got %d", app.SelectionIndex)
	}

	// 1. Arrow navigation
	app.handleIntroKey("down")
	if app.SelectionIndex != 1 {
		t.Errorf("expected SelectionIndex 1 after down, got %d", app.SelectionIndex)
	}
	app.handleIntroKey("j")
	if app.SelectionIndex != 2 {
		t.Errorf("expected SelectionIndex 2 after j, got %d", app.SelectionIndex)
	}
	app.handleIntroKey("up")
	if app.SelectionIndex != 1 {
		t.Errorf("expected SelectionIndex 1 after up, got %d", app.SelectionIndex)
	}
	app.handleIntroKey("k")
	if app.SelectionIndex != 0 {
		t.Errorf("expected SelectionIndex 0 after k, got %d", app.SelectionIndex)
	}
	// Bound check: up at 0
	app.handleIntroKey("up")
	if app.SelectionIndex != 0 {
		t.Errorf("expected SelectionIndex clamped at 0, got %d", app.SelectionIndex)
	}

	// 2. Enter on Search (index 1)
	app.SelectionIndex = 1
	app.handleIntroKey("enter")
	if app.Mode != components.ViewModeSearchInput {
		t.Errorf("expected ViewModeSearchInput after Enter on item 1, got %v", app.Mode)
	}

	// 3. Enter on Library (index 0)
	app.SetMode(components.ViewModeIntro)
	app.SelectionIndex = 0
	app.handleIntroKey("enter")
	if app.Mode != components.ViewModeLibrary {
		t.Errorf("expected ViewModeLibrary after Enter on item 0, got %v", app.Mode)
	}

	// 4. Enter on PlaylistBrowser (index 2)
	app.SetMode(components.ViewModeIntro)
	app.SelectionIndex = 2
	app.handleIntroKey("enter")
	if app.Mode != components.ViewModePlaylistBrowser {
		t.Errorf("expected ViewModePlaylistBrowser after Enter on item 2, got %v", app.Mode)
	}

	// 5. Enter on Help (index 4)
	app.SetMode(components.ViewModeIntro)
	app.SelectionIndex = 4
	app.handleIntroKey("enter")
	if !app.ShowHelpModal {
		t.Errorf("expected ShowHelpModal true after Enter on item 4")
	}

	// 6. Enter on Quit (index 5)
	app.SetMode(components.ViewModeIntro)
	app.SelectionIndex = 5
	app.handleIntroKey("enter")
	if !app.ShowConfirmQuit {
		t.Errorf("expected ShowConfirmQuit true after Enter on item 5")
	}

	// 7. Theme cycling
	app.SetMode(components.ViewModeIntro)
	origTheme := app.Theme.Name
	app.handleIntroKey("t")
	if app.Theme.Name == origTheme && len(theme.BuiltinThemes) > 1 {
		t.Errorf("expected theme to cycle, got %s", app.Theme.Name)
	}
}

func TestEscNavigationToHomeWhenIdle(t *testing.T) {
	mock := player.NewMockPlayer()
	app := NewAppModel(mock)

	// 1. ESC from Library when idle -> ViewModeIntro
	app.SetMode(components.ViewModeLibrary)
	app.handleLibraryKey("esc")
	if app.Mode != components.ViewModeIntro {
		t.Errorf("expected ViewModeIntro when ESC from Library while idle, got %v", app.Mode)
	}

	// 2. ESC from SearchInput when idle -> ViewModeIntro
	app.SetMode(components.ViewModeSearchInput)
	app.handleSearchInputKey(tea.KeyMsg{Type: tea.KeyEsc})
	if app.Mode != components.ViewModeIntro {
		t.Errorf("expected ViewModeIntro when ESC from SearchInput while idle, got %v", app.Mode)
	}

	// 3. ESC from SearchResults when idle -> ViewModeIntro
	app.SetMode(components.ViewModeSearchResults)
	app.handleSearchResultsKey("esc")
	if app.Mode != components.ViewModeIntro {
		t.Errorf("expected ViewModeIntro when ESC from SearchResults while idle, got %v", app.Mode)
	}

	// 4. ESC from PlaylistBrowser when idle -> ViewModeIntro
	app.SetMode(components.ViewModePlaylistBrowser)
	app.handlePlaylistBrowserKey("esc")
	if app.Mode != components.ViewModeIntro {
		t.Errorf("expected ViewModeIntro when ESC from PlaylistBrowser while idle, got %v", app.Mode)
	}

	// 5. ESC from Queue when idle -> ViewModeIntro
	app.SetMode(components.ViewModeQueue)
	app.handleQueueKey("esc")
	if app.Mode != components.ViewModeIntro {
		t.Errorf("expected ViewModeIntro when ESC from Queue while idle, got %v", app.Mode)
	}

	// 6. When music IS playing, ESC from Library -> ViewModePlayback
	_ = mock.Load("/path/to/song.mp3", "replace")
	app.SetMode(components.ViewModeLibrary)
	app.handleLibraryKey("esc")
	if app.Mode != components.ViewModePlayback {
		t.Errorf("expected ViewModePlayback when ESC from Library while playing, got %v", app.Mode)
	}
}

func TestAutoplayAndVolumePersistence(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	mock := player.NewMockPlayer()
	app := NewAppModel(mock)

	// Verify defaults
	if !app.Autoplay {
		t.Errorf("expected default Autoplay to be true, got %v", app.Autoplay)
	}
	if app.Player.Volume() != 100 {
		t.Errorf("expected default Player volume to be 100, got %d", app.Player.Volume())
	}

	// Toggle autoplay via key in Playback view
	app.SetMode(components.ViewModePlayback)
	_ = app.handlePlaybackKey("o")
	if app.Autoplay {
		t.Errorf("expected Autoplay to be toggled off, got %v", app.Autoplay)
	}

	// Adjust volume via keys
	_ = app.handlePlaybackKey("-") // 100 - 5 = 95
	if app.Player.Volume() != 95 {
		t.Errorf("expected Player volume 95, got %d", app.Player.Volume())
	}
	_ = app.handlePlaybackKey("-") // 95 - 5 = 90
	if app.Player.Volume() != 90 {
		t.Errorf("expected Player volume 90, got %d", app.Player.Volume())
	}

	// Recreate app model from the same environment (simulating app restart)
	mock2 := player.NewMockPlayer()
	app2 := NewAppModel(mock2)

	if app2.Autoplay != false {
		t.Errorf("expected restored Autoplay to be false, got %v", app2.Autoplay)
	}
	if app2.Player.Volume() != 90 {
		t.Errorf("expected restored Player volume to be 90, got %d", app2.Player.Volume())
	}

	// Toggle autoplay back on and increase volume
	_ = app2.handlePlaybackKey("o")
	if !app2.Autoplay {
		t.Errorf("expected Autoplay toggled back on, got %v", app2.Autoplay)
	}
	_ = app2.handlePlaybackKey("+") // 90 + 5 = 95
	if app2.Player.Volume() != 95 {
		t.Errorf("expected Player volume 95, got %d", app2.Player.Volume())
	}

	// Recreate again
	mock3 := player.NewMockPlayer()
	app3 := NewAppModel(mock3)

	if !app3.Autoplay {
		t.Errorf("expected restored Autoplay to be true, got %v", app3.Autoplay)
	}
	if app3.Player.Volume() != 95 {
		t.Errorf("expected restored Player volume to be 95, got %d", app3.Player.Volume())
	}
}

func TestPlaylistAddAndSearchNavigationWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	mock := player.NewMockPlayer()
	app := NewAppModel(mock)

	// Create a test playlist
	_ = app.PlaylistManager.CreatePlaylist("MyFavorites")
	app.Playlists = app.PlaylistManager.ListPlaylists()

	// 1. Initially no search results, pressing 's' on intro opens SearchInput
	app.SetMode(components.ViewModeIntro)
	_ = app.handleIntroKey("s")
	if app.Mode != components.ViewModeSearchInput {
		t.Errorf("expected ViewModeSearchInput when no search results, got %v", app.Mode)
	}

	// 2. Populate SearchResults (simulating successful search)
	app.SearchResults = []search.SearchResult{
		{Title: "Song 1", URL: "https://example.com/1", Duration: "03:00", Uploader: "Artist 1"},
		{Title: "Song 2", URL: "https://example.com/2", Duration: "04:00", Uploader: "Artist 2"},
	}
	app.SetMode(components.ViewModeSearchResults)

	// 3. In SearchResults, pressing 's' opens SearchInput for new search
	_ = app.handleSearchResultsKey("s")
	if app.Mode != components.ViewModeSearchInput {
		t.Errorf("expected ViewModeSearchInput on 's' in SearchResults, got %v", app.Mode)
	}

	// 4. In SearchInput, pressing ESC returns to existing SearchResults
	_ = app.handleSearchInputKey(tea.KeyMsg{Type: tea.KeyEsc})
	if app.Mode != components.ViewModeSearchResults {
		t.Errorf("expected ViewModeSearchResults on ESC from SearchInput with results, got %v", app.Mode)
	}

	// 5. Select Song 1 and play it (Enter) -> shifts results to queue, enters Playback
	app.SelectionIndex = 0
	_ = app.handleSearchResultsKey("enter")
	if app.Mode != components.ViewModePlayback {
		t.Errorf("expected ViewModePlayback after playing search result, got %v", app.Mode)
	}
	if len(app.PlayQueue) != 2 {
		t.Fatalf("expected PlayQueue to have 2 items, got %d", len(app.PlayQueue))
	}

	// 6. From Playback, pressing 's' returns to active SearchResults (nothing lost!)
	_ = app.handlePlaybackKey("s")
	if app.Mode != components.ViewModeSearchResults {
		t.Errorf("expected ViewModeSearchResults on 's' from Playback, got %v", app.Mode)
	}

	// 7. From SearchResults, press 'a' on Song 2 to add to playlist -> returns to SearchResults after adding!
	app.SelectionIndex = 1
	_ = app.handleSearchResultsKey("a")
	if app.Mode != components.ViewModePlaylistSelectAdd {
		t.Errorf("expected ViewModePlaylistSelectAdd, got %v", app.Mode)
	}
	if app.SongToAdd.Title != "Song 2" {
		t.Errorf("expected SongToAdd 'Song 2', got %q", app.SongToAdd.Title)
	}
	app.SelectionIndex = 0 // select MyFavorites
	_ = app.handlePlaylistSelectKey("enter")
	if app.Mode != components.ViewModeSearchResults {
		t.Errorf("expected return to ViewModeSearchResults after adding from search, got %v", app.Mode)
	}

	// 8. Go back to Playback screen, press 'a' to add currently playing song (Song 1)
	app.SetMode(components.ViewModePlayback)
	_ = app.handlePlaybackKey("a")
	if app.Mode != components.ViewModePlaylistSelectAdd {
		t.Errorf("expected ViewModePlaylistSelectAdd on 'a' in Playback, got %v", app.Mode)
	}
	if app.SongToAdd.Title != "Song 1" {
		t.Errorf("expected SongToAdd 'Song 1', got %q", app.SongToAdd.Title)
	}
	// Select MyFavorites
	app.SelectionIndex = 0
	_ = app.handlePlaylistSelectKey("enter")
	if app.Mode != components.ViewModePlayback {
		t.Errorf("expected return to ViewModePlayback after adding from playback, got %v", app.Mode)
	}

	// 9. Go to Queue (C key), select Song 2, press 'a' to add to playlist -> returns to Queue!
	app.SetMode(components.ViewModeQueue)
	app.SelectionIndex = 1
	_ = app.handleQueueKey("a")
	if app.Mode != components.ViewModePlaylistSelectAdd {
		t.Errorf("expected ViewModePlaylistSelectAdd on 'a' in Queue, got %v", app.Mode)
	}
	if app.SongToAdd.Title != "Song 2" {
		t.Errorf("expected SongToAdd 'Song 2', got %q", app.SongToAdd.Title)
	}
	app.SelectionIndex = 0
	_ = app.handlePlaylistSelectKey("enter")
	if app.Mode != components.ViewModeQueue {
		t.Errorf("expected return to ViewModeQueue after adding from queue, got %v", app.Mode)
	}

	// 10. Verify playlist contains added songs
	songs := app.PlaylistManager.GetPlaylistSongs("MyFavorites")
	if len(songs) < 2 {
		t.Errorf("expected at least 2 songs in MyFavorites, got %d", len(songs))
	}
}

func TestPlaylistRemoveSongConfirmation(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	mock := player.NewMockPlayer()
	app := NewAppModel(mock)

	// 1. Create playlist with 2 songs
	plName := "RockHits"
	_ = app.PlaylistManager.CreatePlaylist(plName)
	_ = app.PlaylistManager.AddSongToPlaylist(plName, playlist.PlaylistSong{Title: "Bohemian Rhapsody", URL: "/path/1.mp3", Duration: "05:55"})
	_ = app.PlaylistManager.AddSongToPlaylist(plName, playlist.PlaylistSong{Title: "Hotel California", URL: "/path/2.mp3", Duration: "06:30"})

	app.CurrentPlaylistName = plName
	app.PlaylistSongs = app.PlaylistManager.GetPlaylistSongs(plName)
	app.SetMode(components.ViewModePlaylistView)
	app.SelectionIndex = 0

	// 2. Press 'd' to remove "Bohemian Rhapsody" -> should NOT immediately remove!
	_ = app.handlePlaylistSongsKey("d")
	if !app.ShowConfirmDialog {
		t.Fatalf("expected ShowConfirmDialog true on 'd', got false")
	}
	if app.ConfirmDialogTitle != "Remove Song" {
		t.Errorf("expected ConfirmDialogTitle 'Remove Song', got %q", app.ConfirmDialogTitle)
	}
	if app.ConfirmDialogSelection != 1 {
		t.Errorf("expected default selection 1 (NO), got %d", app.ConfirmDialogSelection)
	}

	// 3. User cancels via ESC
	_ = app.handleConfirmDialogKey(tea.KeyMsg{Type: tea.KeyEsc})
	if app.ShowConfirmDialog {
		t.Errorf("expected ShowConfirmDialog false after ESC")
	}
	// Verify songs were not removed
	if len(app.PlaylistSongs) != 2 {
		t.Fatalf("expected 2 songs after cancel, got %d", len(app.PlaylistSongs))
	}

	// 4. Press 'd' again, then cancel via 'n'
	_ = app.handlePlaylistSongsKey("d")
	_ = app.handleConfirmDialogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if len(app.PlaylistSongs) != 2 {
		t.Fatalf("expected 2 songs after 'n', got %d", len(app.PlaylistSongs))
	}

	// 5. Press 'd', toggle to YES (0), and press enter
	_ = app.handlePlaylistSongsKey("d")
	// toggle selection from 1 (NO) to 0 (YES)
	_ = app.handleConfirmDialogKey(tea.KeyMsg{Type: tea.KeyLeft})
	if app.ConfirmDialogSelection != 0 {
		t.Fatalf("expected selection 0 (YES) after left, got %d", app.ConfirmDialogSelection)
	}
	_ = app.handleConfirmDialogKey(tea.KeyMsg{Type: tea.KeyEnter})
	if app.ShowConfirmDialog {
		t.Errorf("expected ShowConfirmDialog false after Enter")
	}

	// Verify song was removed
	if len(app.PlaylistSongs) != 1 {
		t.Fatalf("expected 1 song remaining after confirm, got %d", len(app.PlaylistSongs))
	}
	if app.PlaylistSongs[0].Title != "Hotel California" {
		t.Errorf("expected remaining song 'Hotel California', got %q", app.PlaylistSongs[0].Title)
	}

	// 6. Test quick confirm with 'y'
	app.SelectionIndex = 0
	_ = app.handlePlaylistSongsKey("d")
	_ = app.handleConfirmDialogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if len(app.PlaylistSongs) != 0 {
		t.Errorf("expected 0 songs remaining after 'y' confirm, got %d", len(app.PlaylistSongs))
	}
}

func TestInViewLiveFiltering(t *testing.T) {
	app := NewAppModel(player.NewMockPlayer())

	app.LibraryItems = []library.LibraryItem{
		{Name: "Bohemian Rhapsody.mp3", IsDirectory: false},
		{Name: "Stairway to Heaven.flac", IsDirectory: false},
		{Name: "Hotel California.opus", IsDirectory: false},
		{Name: "Rock Classics", IsDirectory: true},
	}

	app.Playlists = []playlist.Playlist{
		{Name: "Chill Beats", SongCount: 5},
		{Name: "Classic Rock", SongCount: 12},
		{Name: "Workout Pump", SongCount: 8},
	}

	app.PlaylistSongs = []playlist.PlaylistSong{
		{Title: "Comfortably Numb", Duration: "06:22"},
		{Title: "Wish You Were Here", Duration: "05:34"},
		{Title: "Time", Duration: "06:53"},
	}

	app.PlayQueue = []playlist.PlaylistSong{
		{Title: "Track A", Duration: "03:00"},
		{Title: "Track B", Duration: "04:00"},
		{Title: "Echo Track", Duration: "05:00"},
	}

	// 1. Test Library filtering
	app.SetMode(components.ViewModeLibrary)
	if app.FilterMode {
		t.Fatal("expected FilterMode to be false on SetMode")
	}

	// Activate filter with '/'
	app.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.FilterMode {
		t.Fatal("expected FilterMode to be true after pressing '/'")
	}

	// Type 'rock'
	for _, r := range "rock" {
		app.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if app.FilterQuery != "rock" {
		t.Errorf("expected FilterQuery 'rock', got '%s'", app.FilterQuery)
	}

	filteredLib := app.getFilteredLibraryItems()
	if len(filteredLib) != 1 || filteredLib[0].Name != "Rock Classics" {
		t.Errorf("expected 1 match ('Rock Classics'), got %d", len(filteredLib))
	}

	// Press backspace
	app.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if app.FilterQuery != "roc" {
		t.Errorf("expected FilterQuery 'roc', got '%s'", app.FilterQuery)
	}

	// Press Esc to clear
	app.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if app.FilterMode || app.FilterQuery != "" {
		t.Errorf("expected filter cleared on Esc, got mode=%v, query=%s", app.FilterMode, app.FilterQuery)
	}
	if len(app.getFilteredLibraryItems()) != 4 {
		t.Errorf("expected 4 unfiltered items, got %d", len(app.getFilteredLibraryItems()))
	}

	// 2. Test Playlist Songs filtering
	app.SetMode(components.ViewModePlaylistView)
	app.FilterMode = true
	app.FilterQuery = "wish"
	filteredSongs := app.getFilteredPlaylistSongs()
	if len(filteredSongs) != 1 || filteredSongs[0].Title != "Wish You Were Here" {
		t.Errorf("expected 1 match ('Wish You Were Here'), got %d", len(filteredSongs))
	}

	// 3. Test Playlist Browser filtering
	app.SetMode(components.ViewModePlaylistBrowser)
	app.Playlists = []playlist.Playlist{
		{Name: "Chill Beats", SongCount: 5},
		{Name: "Classic Rock", SongCount: 12},
		{Name: "Workout Pump", SongCount: 8},
	}
	app.FilterMode = true
	app.FilterQuery = "chill"
	filteredPlaylists := app.getFilteredPlaylists()
	if len(filteredPlaylists) != 1 || filteredPlaylists[0].Name != "Chill Beats" {
		t.Errorf("expected 1 match ('Chill Beats'), got %d", len(filteredPlaylists))
	}

	// 4. Test Queue filtering
	app.SetMode(components.ViewModeQueue)
	app.FilterMode = true
	app.FilterQuery = "echo"
	filteredQueue := app.getFilteredQueue()
	if len(filteredQueue) != 1 || filteredQueue[0].Title != "Echo Track" {
		t.Errorf("expected 1 match ('Echo Track'), got %d", len(filteredQueue))
	}
}

func TestQueueStreamPrefetch(t *testing.T) {
	app := NewAppModel(player.NewMockPlayer())

	url1 := "https://www.youtube.com/watch?v=song1"
	url2 := "https://www.youtube.com/watch?v=song2"

	// Mock pre-cached stream for song2
	search.DefaultStreamCache.Set(url2, search.StreamInfo{
		Title:     "Song 2",
		Artist:    "Artist 2",
		Duration:  200,
		StreamURL: "https://stream.googlevideo.com/playback2",
	})

	app.PlayQueue = []playlist.PlaylistSong{
		{Title: "Song 1", URL: url1, Duration: "03:00"},
		{Title: "Song 2", URL: url2, Duration: "03:20"},
	}
	app.QueueIndex = 0

	// Check prefetchUpcomingTracks does not panic and detects next track
	app.prefetchUpcomingTracks()

	if !search.DefaultStreamCache.Has(url2) {
		t.Errorf("expected stream cache to have %s", url2)
	}

	// Verify StartTrackPlayback uses cached stream URL directly
	_ = app.StartTrackPlayback("Song 2", url2, "03:20", "Artist 2")
	mock, ok := app.Player.(*player.MockPlayer)
	if ok {
		if mock.GetPath() != "https://stream.googlevideo.com/playback2" {
			t.Errorf("expected Player to load cached StreamURL, got %q", mock.GetPath())
		}
	}
}

