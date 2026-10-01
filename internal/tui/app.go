package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"vibe-fi/internal/config"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/tui/views"
	"vibe-fi/internal/tui/visualizer"
	"vibe-fi/internal/utils/mem"
	"vibe-fi/internal/utils/net"
	"vibe-fi/internal/utils/stringutil"
)

// AppModel is the top-level Bubble Tea model orchestrating all views and state.
type AppModel struct {
	Player          player.AudioPlayer
	Library         *library.Library
	PlaylistManager *playlist.PlaylistManager
	LyricsManager   *lyrics.LyricsManager
	Visualizer      *visualizer.Visualizer
	trimCounter     int

	Width  int
	Height int

	Mode    components.ViewMode
	Theme   theme.Theme
	Styles  theme.Styles
	VizMode visualizer.VisualizerMode

	Autoplay      bool
	StatusMessage string

	CurrentPath  string
	LibraryItems []library.LibraryItem

	SearchQuery   string
	SearchResults []search.SearchResult

	Playlists            []playlist.Playlist
	PlaylistSongs        []playlist.PlaylistSong
	PreviewSongs         []playlist.PlaylistSong
	CurrentPlaylistName  string
	PlayingPlaylistName  string
	IsPlayingFromPlaylist bool

	PlayQueue  []playlist.PlaylistSong
	QueueIndex int

	SelectionIndex int
	ScrollOffset   int

	LyricsData         lyrics.LyricsData
	LyricsScrollOffset int
	LyricsAutoScroll   bool
	LyricsRequestID    uint64
	CurrentLyricsTitle string
	LastPlayedPath     string

	// Modals
	ShowConfirmQuit      bool
	ConfirmQuitSelection int // 0: YES, 1: NO

	ShowInputPrompt     bool
	InputPromptTitle    string
	InputPromptText     string
	InputPromptCallback func(text string) tea.Cmd

	SongToMoveIndex  int
	SongToMoveOrigin string
	SongToAdd        playlist.PlaylistSong

	Quitting bool
}

// NewAppModel creates a fully initialized AppModel.
func NewAppModel(p player.AudioPlayer) *AppModel {
	th := theme.GetTheme(config.DefaultTheme)
	lib := library.NewLibrary()
	homeDir := lib.GetHomeMusicDir()
	items, _ := lib.ListDirectory(homeDir)

	return &AppModel{
		Player:               p,
		Library:              lib,
		PlaylistManager:      playlist.NewPlaylistManager(),
		LyricsManager:        lyrics.NewLyricsManager(),
		Visualizer:           visualizer.NewVisualizer(),
		Mode:                 components.ViewModeIntro,
		Theme:                th,
		Styles:               theme.MakeStyles(th),
		VizMode:              visualizer.ModeCavaWave,
		CurrentPath:          homeDir,
		LibraryItems:         items,
		QueueIndex:           -1,
		LyricsAutoScroll:     true,
		ConfirmQuitSelection: 1, // Default NO
	}
}

// Init triggers initial commands and starts the 30 FPS tick loop.
func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
	)
}

func (m *AppModel) tickCmd() tea.Cmd {
	return tea.Tick(33*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *AppModel) statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

// ShowStatus displays a temporary alert for 3 seconds.
func (m *AppModel) ShowStatus(msg string) tea.Cmd {
	m.StatusMessage = msg
	return m.statusClearCmd()
}

// SetMode switches active view and resets selection.
func (m *AppModel) SetMode(newMode components.ViewMode) {
	m.Mode = newMode
	m.SelectionIndex = 0
	m.ScrollOffset = 0

	if m.Mode == components.ViewModePlaylistBrowser {
		m.Playlists = m.PlaylistManager.ListPlaylists()
		m.updatePreviewSongs()
	}
}

func (m *AppModel) updatePreviewSongs() {
	if len(m.Playlists) > 0 && m.SelectionIndex >= 0 && m.SelectionIndex < len(m.Playlists) {
		m.PreviewSongs = m.PlaylistManager.GetPlaylistSongs(m.Playlists[m.SelectionIndex].Name)
	} else {
		m.PreviewSongs = nil
	}
}

// StartTrackPlayback loads and plays a track, updating queue, metadata, and lyrics.
func (m *AppModel) StartTrackPlayback(title, url, duration, artistHint string) tea.Cmd {
	m.LastPlayedPath = url
	m.LyricsScrollOffset = 0
	m.LyricsAutoScroll = true
	m.CurrentLyricsTitle = title

	_ = m.Player.Load(url, "replace")
	if title != "" {
		_ = m.Player.SetProperty("force-media-title", title)
	}
	_ = m.Player.Play()

	m.saveCurrentState()
	return m.fetchLyricsCmd(title, url, stringutil.ParseDuration(duration), artistHint)
}

func (m *AppModel) fetchLyricsCmd(title, url string, duration float64, artistHint string) tea.Cmd {
	m.LyricsRequestID++
	reqID := m.LyricsRequestID
	mgr := m.LyricsManager

	return func() tea.Msg {
		artist, track := stringutil.CleanTrackTitle(title)
		if artistHint != "" && artist == "" {
			artist = artistHint
		}
		data, err := mgr.FetchLyrics(artist, track, duration)
		return LyricsFetchedMsg{
			RequestID: reqID,
			Title:     title,
			Artist:    artist,
			Data:      data,
			Err:       err,
		}
	}
}

func (m *AppModel) searchYoutubeCmd(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := search.SearchYouTube(query, 10)
		return SearchResultsMsg{
			Query:   query,
			Results: results,
			Err:     err,
		}
	}
}

func (m *AppModel) resolveStreamCmd(url string) tea.Cmd {
	return func() tea.Msg {
		info, err := search.ResolveStreamInfo(url)
		return StreamResolvedMsg{
			URL:  url,
			Info: info,
			Err:  err,
		}
	}
}

func (m *AppModel) saveCurrentState() {
	st := config.SessionState{
		Path:       m.LastPlayedPath,
		Title:      m.CurrentLyricsTitle,
		Position:   m.Player.Position(),
		Volume:     m.Player.Volume(),
		Index:      m.QueueIndex,
		Theme:      m.Theme.Name,
		Visualizer: int(m.VizMode),
		Autoplay:   m.Autoplay,
	}
	if m.IsPlayingFromPlaylist {
		st.Playlist = m.PlayingPlaylistName
	}
	_ = config.SaveState(st)
}

// LoadState restores saved session settings.
func (m *AppModel) LoadState() tea.Cmd {
	st, err := config.LoadState()
	if err != nil {
		return m.ShowStatus("No saved session found.")
	}

	if st.Theme != "" {
		m.Theme = theme.GetTheme(st.Theme)
		m.Styles = theme.MakeStyles(m.Theme)
	}
	m.VizMode = visualizer.VisualizerMode(st.Visualizer)
	m.Autoplay = st.Autoplay
	if st.Volume > 0 {
		_ = m.Player.SetVolume(st.Volume)
	}

	if st.Path != "" {
		if stringutil.IsURL(st.Path) && !net.IsOnline() {
			return m.ShowStatus("Network unavailable: cannot resume online track.")
		}

		m.LastPlayedPath = st.Path
		m.CurrentLyricsTitle = st.Title
		m.QueueIndex = st.Index

		if st.Playlist != "" {
			m.IsPlayingFromPlaylist = true
			m.PlayingPlaylistName = st.Playlist
			m.PlayQueue = m.PlaylistManager.GetPlaylistSongs(st.Playlist)
		}

		if st.Position > 0 {
			_ = m.Player.SetProperty("start", fmt.Sprintf("%.2f", st.Position))
		}
		_ = m.Player.Load(st.Path, "replace")
		_ = m.Player.SetProperty("start", "0")
		if st.Title != "" {
			_ = m.Player.SetProperty("force-media-title", st.Title)
		}
		_ = m.Player.Play()

		m.SetMode(components.ViewModePlayback)
		cmd1 := m.ShowStatus("Resuming session...")
		cmd2 := m.fetchLyricsCmd(st.Title, st.Path, 0, "")
		return tea.Batch(cmd1, cmd2)
	}

	return m.ShowStatus("No previous track found in session.")
}

func (m *AppModel) playNext() tea.Cmd {
	if len(m.PlayQueue) == 0 {
		return nil
	}
	nextIdx := m.QueueIndex + 1
	if nextIdx < len(m.PlayQueue) {
		song := m.PlayQueue[nextIdx]
		if stringutil.IsURL(song.URL) && !net.IsOnline() {
			return m.ShowStatus("Network unavailable: Paused at " + song.Title)
		}
		m.QueueIndex = nextIdx
		statusCmd := m.ShowStatus("Playing: " + song.Title)
		playCmd := m.StartTrackPlayback(song.Title, song.URL, song.Duration, m.PlayingPlaylistName)
		return tea.Batch(statusCmd, playCmd)
	}
	m.QueueIndex = -1
	return m.ShowStatus("Reached end of queue.")
}

func (m *AppModel) playPrevious() tea.Cmd {
	if m.Player.Position() > 3.0 {
		_ = m.Player.Seek(-m.Player.Position())
		return nil
	}
	if m.QueueIndex > 0 && m.QueueIndex <= len(m.PlayQueue) {
		prevIdx := m.QueueIndex - 1
		song := m.PlayQueue[prevIdx]
		if stringutil.IsURL(song.URL) && !net.IsOnline() {
			return m.ShowStatus("Network unavailable: Cannot play " + song.Title)
		}
		m.QueueIndex = prevIdx
		statusCmd := m.ShowStatus("Playing previous: " + song.Title)
		playCmd := m.StartTrackPlayback(song.Title, song.URL, song.Duration, m.PlayingPlaylistName)
		return tea.Batch(statusCmd, playCmd)
	}
	_ = m.Player.Seek(-m.Player.Position())
	return nil
}

// Update processes incoming messages and keyboard events.
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case TickMsg:
		if m.Quitting {
			return m, nil
		}
		m.trimCounter++
		if m.trimCounter >= 150 {
			m.trimCounter = 0
			mem.PeriodicTrim()
		}

		// Poll libmpv events
		m.Player.PollEvents()

		// Invariant 6: Natural EOF check for autoplay
		if m.Player.ConsumeTrackFinished() {
			if m.Autoplay {
				nextCmd := m.playNext()
				return m, tea.Batch(nextCmd, m.tickCmd())
			}
		}

		if m.Player.ConsumePlaybackError() {
			errStr := m.Player.GetLastError()
			if errStr == "" {
				errStr = "Playback error occurred"
			}
			statusCmd := m.ShowStatus("[error] " + errStr)
			return m, tea.Batch(statusCmd, m.tickCmd())
		}

		return m, m.tickCmd()

	case ClearStatusMsg:
		m.StatusMessage = ""
		return m, nil

	case LyricsFetchedMsg:
		if msg.RequestID == m.LyricsRequestID {
			m.LyricsData = msg.Data
		}
		return m, nil

	case SearchResultsMsg:
		if msg.Err != nil {
			return m, m.ShowStatus("Search failed: " + msg.Err.Error())
		}
		m.SearchResults = msg.Results
		m.SetMode(components.ViewModeSearchResults)
		return m, nil

	case StreamResolvedMsg:
		if msg.Err != nil || msg.Info.StreamURL == "" {
			return m, m.ShowStatus("Failed to resolve stream URL.")
		}
		displayTitle := msg.Info.Title
		if msg.Info.Artist != "" && !strings.Contains(displayTitle, " - ") {
			displayTitle = msg.Info.Artist + " - " + displayTitle
		}
		if displayTitle == "" {
			displayTitle = msg.URL
		}
		m.IsPlayingFromPlaylist = false
		m.PlayingPlaylistName = ""
		m.SetMode(components.ViewModePlayback)
		statusCmd := m.ShowStatus("Playing: " + displayTitle)
		playCmd := m.StartTrackPlayback(displayTitle, msg.Info.StreamURL, stringutil.FormatDuration(msg.Info.Duration), msg.Info.Artist)
		return m, tea.Batch(statusCmd, playCmd)

	case MprisActionMsg:
		return m, m.handleMprisAction(msg.Action)

	case tea.KeyMsg:
		// Modal Key Handling
		if m.ShowConfirmQuit {
			return m, m.handleConfirmQuitKey(msg)
		}
		if m.ShowInputPrompt {
			return m, m.handleInputPromptKey(msg)
		}

		// View Key Handling
		return m, m.handleKey(msg)
	}

	return m, nil
}

func (m *AppModel) handleMprisAction(action MprisAction) tea.Cmd {
	switch action {
	case MprisPlayPause:
		_ = m.Player.TogglePause()
	case MprisPlay:
		_ = m.Player.Play()
	case MprisPause:
		_ = m.Player.Pause()
	case MprisNext:
		return m.playNext()
	case MprisPrevious:
		return m.playPrevious()
	case MprisStop:
		_ = m.Player.Stop()
	}
	return nil
}

func (m *AppModel) handleConfirmQuitKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "left", "right", "h", "l", "tab":
		m.ConfirmQuitSelection = 1 - m.ConfirmQuitSelection
	case "enter":
		if m.ConfirmQuitSelection == 0 { // YES
			m.saveCurrentState()
			_ = m.Player.Stop()
			m.Quitting = true
			return tea.Quit
		}
		m.ShowConfirmQuit = false
	case "esc", "n", "N":
		m.ShowConfirmQuit = false
	case "y", "Y":
		m.saveCurrentState()
		_ = m.Player.Stop()
		m.Quitting = true
		return tea.Quit
	}
	return nil
}

func (m *AppModel) handleInputPromptKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.ShowInputPrompt = false
		m.InputPromptText = ""
	case "enter":
		text := strings.TrimSpace(m.InputPromptText)
		m.ShowInputPrompt = false
		m.InputPromptText = ""
		if m.InputPromptCallback != nil {
			return m.InputPromptCallback(text)
		}
	case "backspace":
		if len(m.InputPromptText) > 0 {
			m.InputPromptText = m.InputPromptText[:len(m.InputPromptText)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.InputPromptText += string(msg.Runes)
		}
	}
	return nil
}

func (m *AppModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if msg.Type == tea.KeyCtrlC {
		m.saveCurrentState()
		_ = m.Player.Stop()
		m.Quitting = true
		return tea.Quit
	}

	keyStr := msg.String()

	// Global Hotkeys when in Playback mode
	if m.Mode == components.ViewModePlayback {
		switch keyStr {
		case "esc", "q", "Q":
			m.ShowConfirmQuit = true
			m.ConfirmQuitSelection = 1
			return nil
		case " ":
			_ = m.Player.TogglePause()
			return nil
		case "l", "L":
			m.SetMode(components.ViewModeLibrary)
			return nil
		case "s", "S":
			m.SearchQuery = ""
			m.SetMode(components.ViewModeSearchInput)
			return nil
		case "p", "P":
			m.Playlists = m.PlaylistManager.ListPlaylists()
			m.SetMode(components.ViewModePlaylistBrowser)
			return nil
		case "c", "C":
			m.SetMode(components.ViewModeQueue)
			return nil
		case "r", "R":
			if m.LastPlayedPath != "" {
				if stringutil.IsURL(m.LastPlayedPath) && !net.IsOnline() {
					return m.ShowStatus("Network unavailable: cannot stream track.")
				}
				_ = m.Player.Load(m.LastPlayedPath, "replace")
				_ = m.Player.Play()
				m.LyricsScrollOffset = 0
				m.LyricsAutoScroll = true
				statusCmd := m.ShowStatus("Replaying...")
				fetchCmd := m.fetchLyricsCmd(m.CurrentLyricsTitle, m.LastPlayedPath, 0, "")
				return tea.Batch(statusCmd, fetchCmd)
			}
			return nil
		case "left":
			_ = m.Player.Seek(-5.0)
			return nil
		case "right":
			_ = m.Player.Seek(5.0)
			return nil
		case "+", "=":
			_ = m.Player.SetVolume(m.Player.Volume() + 5)
			return nil
		case "-", "_":
			_ = m.Player.SetVolume(m.Player.Volume() - 5)
			return nil
		case "o", "O":
			m.Autoplay = !m.Autoplay
			autoStr := "OFF"
			if m.Autoplay {
				autoStr = "ON"
			}
			return m.ShowStatus("Autoplay: " + autoStr)
		case "t", "T":
			m.Theme = theme.CycleTheme(m.Theme.Name)
			m.Styles = theme.MakeStyles(m.Theme)
			m.saveCurrentState()
			return m.ShowStatus("Theme: " + m.Theme.Name)
		case "v", "V":
			m.VizMode = (m.VizMode + 1) % 3
			m.saveCurrentState()
			return m.ShowStatus("Visualizer: " + m.VizMode.String())
		case "u", "U":
			m.ShowInputPrompt = true
			m.InputPromptTitle = "Paste YouTube URL"
			m.InputPromptText = ""
			m.InputPromptCallback = func(text string) tea.Cmd {
				if text == "" {
					return nil
				}
				if !net.IsOnline() {
					return m.ShowStatus("Network unavailable.")
				}
				statusCmd := m.ShowStatus("Resolving stream URL...")
				resolveCmd := m.resolveStreamCmd(text)
				return tea.Batch(statusCmd, resolveCmd)
			}
			return nil
		case ">", ".", "n", "N":
			return m.playNext()
		case "<", ",", "b", "B":
			return m.playPrevious()
		case "up", "k":
			if m.LyricsScrollOffset > 0 {
				m.LyricsScrollOffset--
			}
			m.LyricsAutoScroll = false
			return nil
		case "down", "j":
			m.LyricsScrollOffset++
			m.LyricsAutoScroll = false
			return nil
		}
	}

	// Mode-specific input handling
	switch m.Mode {

	case components.ViewModeIntro:
		switch keyStr {
		case "enter", "l", "L":
			m.SetMode(components.ViewModeLibrary)
		case "s", "S":
			m.SearchQuery = ""
			m.SetMode(components.ViewModeSearchInput)
		case "p", "P":
			m.SetMode(components.ViewModePlaylistBrowser)
		case "r", "R":
			return m.LoadState()
		case "esc", "q", "Q":
			m.ShowConfirmQuit = true
			m.ConfirmQuitSelection = 1
		}

	case components.ViewModeLibrary:
		listH := m.Height - 8 - 2
		if listH < 1 {
			listH = 1
		}
		switch keyStr {
		case "esc":
			m.SetMode(components.ViewModePlayback)
		case "up", "k":
			if m.SelectionIndex > 0 {
				m.SelectionIndex--
				if m.SelectionIndex < m.ScrollOffset {
					m.ScrollOffset = m.SelectionIndex
				}
			}
		case "down", "j":
			if m.SelectionIndex < len(m.LibraryItems)-1 {
				m.SelectionIndex++
				if m.SelectionIndex >= m.ScrollOffset+listH {
					m.ScrollOffset = m.SelectionIndex - listH + 1
				}
			}
		case "backspace", "h":
			parent := filepath.Dir(m.CurrentPath)
			if parent != m.CurrentPath {
				m.CurrentPath = parent
				m.LibraryItems, _ = m.Library.ListDirectory(parent)
				m.SelectionIndex = 0
				m.ScrollOffset = 0
			}
		case "enter":
			if len(m.LibraryItems) > 0 && m.SelectionIndex < len(m.LibraryItems) {
				item := m.LibraryItems[m.SelectionIndex]
				if item.IsDirectory {
					m.CurrentPath = item.Path
					m.LibraryItems, _ = m.Library.ListDirectory(item.Path)
					m.SelectionIndex = 0
					m.ScrollOffset = 0
				} else {
					title := strings.TrimSuffix(item.Name, filepath.Ext(item.Name))
					m.PlayQueue = []playlist.PlaylistSong{{Title: title, URL: item.Path, Duration: item.Duration}}
					m.QueueIndex = 0
					m.IsPlayingFromPlaylist = false
					m.SetMode(components.ViewModePlayback)
					statusCmd := m.ShowStatus("Playing: " + title)
					playCmd := m.StartTrackPlayback(title, item.Path, item.Duration, "")
					return tea.Batch(statusCmd, playCmd)
				}
			}
		case "a", "A":
			if len(m.LibraryItems) > 0 && m.SelectionIndex < len(m.LibraryItems) {
				item := m.LibraryItems[m.SelectionIndex]
				if !item.IsDirectory {
					title := strings.TrimSuffix(item.Name, filepath.Ext(item.Name))
					m.SongToAdd = playlist.PlaylistSong{Title: title, URL: item.Path, Duration: item.Duration}
					m.Playlists = m.PlaylistManager.ListPlaylists()
					m.SetMode(components.ViewModePlaylistSelectAdd)
				}
			}
		}

	case components.ViewModeSearchInput:
		switch keyStr {
		case "esc":
			m.SetMode(components.ViewModePlayback)
		case "enter":
			q := strings.TrimSpace(m.SearchQuery)
			if q != "" {
				statusCmd := m.ShowStatus("Searching YouTube...")
				searchCmd := m.searchYoutubeCmd(q)
				return tea.Batch(statusCmd, searchCmd)
			}
		case "backspace":
			if len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				m.SearchQuery += string(msg.Runes)
			}
		}

	case components.ViewModeSearchResults:
		listH := m.Height - 8 - 3
		if listH < 1 {
			listH = 1
		}
		switch keyStr {
		case "esc":
			m.SetMode(components.ViewModePlayback)
		case "s", "S":
			m.SearchQuery = ""
			m.SetMode(components.ViewModeSearchInput)
		case "up", "k":
			if m.SelectionIndex > 0 {
				m.SelectionIndex--
				if m.SelectionIndex < m.ScrollOffset {
					m.ScrollOffset = m.SelectionIndex
				}
			}
		case "down", "j":
			if m.SelectionIndex < len(m.SearchResults)-1 {
				m.SelectionIndex++
				if m.SelectionIndex >= m.ScrollOffset+listH {
					m.ScrollOffset = m.SelectionIndex - listH + 1
				}
			}
		case "enter":
			if len(m.SearchResults) > 0 && m.SelectionIndex < len(m.SearchResults) {
				hit := m.SearchResults[m.SelectionIndex]
				var queue []playlist.PlaylistSong
				for _, res := range m.SearchResults {
					queue = append(queue, playlist.PlaylistSong{Title: res.Title, URL: res.URL, Duration: res.Duration})
				}
				m.PlayQueue = queue
				m.QueueIndex = m.SelectionIndex
				m.IsPlayingFromPlaylist = false
				m.PlayingPlaylistName = ""
				m.SetMode(components.ViewModePlayback)
				statusCmd := m.ShowStatus("Playing: " + hit.Title)
				playCmd := m.StartTrackPlayback(hit.Title, hit.URL, hit.Duration, hit.Uploader)
				return tea.Batch(statusCmd, playCmd)
			}
		case "a", "A":
			if len(m.SearchResults) > 0 && m.SelectionIndex < len(m.SearchResults) {
				hit := m.SearchResults[m.SelectionIndex]
				m.SongToAdd = playlist.PlaylistSong{Title: hit.Title, URL: hit.URL, Duration: hit.Duration}
				m.Playlists = m.PlaylistManager.ListPlaylists()
				m.SetMode(components.ViewModePlaylistSelectAdd)
			}
		}

	case components.ViewModePlaylistBrowser:
		listH := m.Height - 8 - 3
		if listH < 1 {
			listH = 1
		}
		switch keyStr {
		case "esc":
			m.SetMode(components.ViewModePlayback)
		case "up", "k":
			if m.SelectionIndex > 0 {
				m.SelectionIndex--
				if m.SelectionIndex < m.ScrollOffset {
					m.ScrollOffset = m.SelectionIndex
				}
				m.updatePreviewSongs()
			}
		case "down", "j":
			if m.SelectionIndex < len(m.Playlists)-1 {
				m.SelectionIndex++
				if m.SelectionIndex >= m.ScrollOffset+listH {
					m.ScrollOffset = m.SelectionIndex - listH + 1
				}
				m.updatePreviewSongs()
			}
		case "enter":
			if len(m.Playlists) > 0 && m.SelectionIndex < len(m.Playlists) {
				m.CurrentPlaylistName = m.Playlists[m.SelectionIndex].Name
				m.PlaylistSongs = m.PlaylistManager.GetPlaylistSongs(m.CurrentPlaylistName)
				m.SetMode(components.ViewModePlaylistView)
			}
		case "n", "N":
			m.ShowInputPrompt = true
			m.InputPromptTitle = "New Playlist Name"
			m.InputPromptText = ""
			m.InputPromptCallback = func(text string) tea.Cmd {
				if text != "" {
					_ = m.PlaylistManager.CreatePlaylist(text)
					m.Playlists = m.PlaylistManager.ListPlaylists()
					m.updatePreviewSongs()
					return m.ShowStatus("Created playlist: " + text)
				}
				return nil
			}
		case "r", "R":
			if len(m.Playlists) > 0 && m.SelectionIndex < len(m.Playlists) {
				oldName := m.Playlists[m.SelectionIndex].Name
				m.ShowInputPrompt = true
				m.InputPromptTitle = "Rename Playlist"
				m.InputPromptText = oldName
				m.InputPromptCallback = func(text string) tea.Cmd {
					if text != "" && text != oldName {
						_ = m.PlaylistManager.RenamePlaylist(oldName, text)
						m.Playlists = m.PlaylistManager.ListPlaylists()
						m.updatePreviewSongs()
						return m.ShowStatus("Renamed to: " + text)
					}
					return nil
				}
			}
		case "d", "D":
			if len(m.Playlists) > 0 && m.SelectionIndex < len(m.Playlists) {
				name := m.Playlists[m.SelectionIndex].Name
				_ = m.PlaylistManager.DeletePlaylist(name)
				m.Playlists = m.PlaylistManager.ListPlaylists()
				if m.SelectionIndex >= len(m.Playlists) && m.SelectionIndex > 0 {
					m.SelectionIndex--
				}
				m.updatePreviewSongs()
				return m.ShowStatus("Deleted playlist: " + name)
			}
		}

	case components.ViewModePlaylistView:
		listH := m.Height - 8 - 3
		if listH < 1 {
			listH = 1
		}
		switch keyStr {
		case "esc":
			m.SetMode(components.ViewModePlaylistBrowser)
		case "up", "k":
			if m.SelectionIndex > 0 {
				m.SelectionIndex--
				if m.SelectionIndex < m.ScrollOffset {
					m.ScrollOffset = m.SelectionIndex
				}
			}
		case "down", "j":
			if m.SelectionIndex < len(m.PlaylistSongs)-1 {
				m.SelectionIndex++
				if m.SelectionIndex >= m.ScrollOffset+listH {
					m.ScrollOffset = m.SelectionIndex - listH + 1
				}
			}
		case "enter":
			if len(m.PlaylistSongs) > 0 && m.SelectionIndex < len(m.PlaylistSongs) {
				s := m.PlaylistSongs[m.SelectionIndex]
				m.PlayQueue = m.PlaylistSongs
				m.QueueIndex = m.SelectionIndex
				m.IsPlayingFromPlaylist = true
				m.PlayingPlaylistName = m.CurrentPlaylistName
				m.SetMode(components.ViewModePlayback)
				statusCmd := m.ShowStatus("Playing: " + s.Title)
				playCmd := m.StartTrackPlayback(s.Title, s.URL, s.Duration, m.CurrentPlaylistName)
				return tea.Batch(statusCmd, playCmd)
			}
		case "d", "D":
			if len(m.PlaylistSongs) > 0 && m.SelectionIndex < len(m.PlaylistSongs) {
				_ = m.PlaylistManager.RemoveSongFromPlaylist(m.CurrentPlaylistName, m.SelectionIndex)
				m.PlaylistSongs = m.PlaylistManager.GetPlaylistSongs(m.CurrentPlaylistName)
				if m.SelectionIndex >= len(m.PlaylistSongs) && m.SelectionIndex > 0 {
					m.SelectionIndex--
				}
				return m.ShowStatus("Song removed from playlist.")
			}
		case "m", "M":
			if len(m.PlaylistSongs) > 0 && m.SelectionIndex < len(m.PlaylistSongs) {
				m.SongToMoveIndex = m.SelectionIndex
				m.SongToMoveOrigin = m.CurrentPlaylistName
				m.Playlists = m.PlaylistManager.ListPlaylists()
				m.SetMode(components.ViewModePlaylistSelectMove)
			}
		}

	case components.ViewModePlaylistSelectAdd, components.ViewModePlaylistSelectMove:
		switch keyStr {
		case "esc":
			if m.Mode == components.ViewModePlaylistSelectMove {
				m.SetMode(components.ViewModePlaylistView)
			} else {
				m.SetMode(components.ViewModePlayback)
			}
		case "up", "k":
			if m.SelectionIndex > 0 {
				m.SelectionIndex--
			}
		case "down", "j":
			if m.SelectionIndex < len(m.Playlists)-1 {
				m.SelectionIndex++
			}
		case "enter":
			if len(m.Playlists) > 0 && m.SelectionIndex < len(m.Playlists) {
				targetName := m.Playlists[m.SelectionIndex].Name
				if m.Mode == components.ViewModePlaylistSelectMove {
					_ = m.PlaylistManager.MoveSong(m.SongToMoveOrigin, m.SongToMoveIndex, targetName)
					m.CurrentPlaylistName = m.SongToMoveOrigin
					m.PlaylistSongs = m.PlaylistManager.GetPlaylistSongs(m.CurrentPlaylistName)
					m.SetMode(components.ViewModePlaylistView)
					return m.ShowStatus("Song moved to " + targetName)
				}
				_ = m.PlaylistManager.AddSongToPlaylist(targetName, m.SongToAdd)
				m.SetMode(components.ViewModePlayback)
				return m.ShowStatus("Added to " + targetName)
			}
		}

	case components.ViewModeQueue:
		listH := m.Height - 8 - 2
		if listH < 1 {
			listH = 1
		}
		switch keyStr {
		case "esc":
			m.SetMode(components.ViewModePlayback)
		case "up", "k":
			if m.SelectionIndex > 0 {
				m.SelectionIndex--
				if m.SelectionIndex < m.ScrollOffset {
					m.ScrollOffset = m.SelectionIndex
				}
			}
		case "down", "j":
			if m.SelectionIndex < len(m.PlayQueue)-1 {
				m.SelectionIndex++
				if m.SelectionIndex >= m.ScrollOffset+listH {
					m.ScrollOffset = m.SelectionIndex - listH + 1
				}
			}
		case "enter":
			if len(m.PlayQueue) > 0 && m.SelectionIndex < len(m.PlayQueue) {
				song := m.PlayQueue[m.SelectionIndex]
				if stringutil.IsURL(song.URL) && !net.IsOnline() {
					return m.ShowStatus("Network unavailable.")
				}
				m.QueueIndex = m.SelectionIndex
				m.SetMode(components.ViewModePlayback)
				statusCmd := m.ShowStatus("Playing: " + song.Title)
				playCmd := m.StartTrackPlayback(song.Title, song.URL, song.Duration, m.PlayingPlaylistName)
				return tea.Batch(statusCmd, playCmd)
			}
		case "d", "D":
			if len(m.PlayQueue) > 0 && m.SelectionIndex < len(m.PlayQueue) {
				m.PlayQueue = append(m.PlayQueue[:m.SelectionIndex], m.PlayQueue[m.SelectionIndex+1:]...)
				if m.QueueIndex == m.SelectionIndex {
					m.QueueIndex = -1
				} else if m.QueueIndex > m.SelectionIndex {
					m.QueueIndex--
				}
				if m.SelectionIndex >= len(m.PlayQueue) && m.SelectionIndex > 0 {
					m.SelectionIndex--
				}
				return m.ShowStatus("Track removed from queue.")
			}
		}

	case components.ViewModeLyrics:
		switch keyStr {
		case "esc", "q", "Q":
			m.SetMode(components.ViewModePlayback)
		case "up", "k":
			if m.LyricsScrollOffset > 0 {
				m.LyricsScrollOffset--
			}
		case "down", "j":
			m.LyricsScrollOffset++
		}
	}

	return nil
}

// View computes the final terminal display frame.
func (m *AppModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	if m.Quitting {
		return ""
	}

	// Active Modal Overlays
	if m.ShowConfirmQuit {
		return components.RenderConfirmQuit(m.Width, m.Height, m.ConfirmQuitSelection, m.Styles)
	}
	if m.ShowInputPrompt {
		return components.RenderInputPrompt(m.Width, m.Height, m.InputPromptTitle, m.InputPromptText, m.Styles)
	}

	// Layout breakdown:
	// statusH = 5
	// helpH   = 3
	// mainH   = height - statusH - helpH
	statusH := 5
	helpH := 3
	mainH := m.Height - statusH - helpH
	if mainH < 6 {
		mainH = 6
	}

	var mainView string
	switch m.Mode {
	case components.ViewModePlayback:
		var newOffset int
		mainView, newOffset = views.RenderPlaybackView(m.Width, mainH, m.Player, m.Visualizer, m.VizMode, m.LyricsData, m.LyricsScrollOffset, m.LyricsAutoScroll, m.Styles)
		m.LyricsScrollOffset = newOffset

	case components.ViewModeLibrary:
		mainView = views.RenderLibraryView(m.Width, mainH, m.CurrentPath, m.LibraryItems, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModeSearchInput:
		mainView = views.RenderSearchInput(m.Width, mainH, m.SearchQuery, m.Styles)

	case components.ViewModeSearchResults:
		mainView = views.RenderSearchResults(m.Width, mainH, m.SearchResults, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModePlaylistBrowser:
		mainView = views.RenderPlaylistsBrowser(m.Width, mainH, m.Playlists, m.PreviewSongs, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModePlaylistView:
		mainView = views.RenderPlaylistSongsView(m.Width, mainH, m.CurrentPlaylistName, m.PlaylistSongs, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModePlaylistSelectAdd:
		mainView = views.RenderPlaylistSelectDialog(m.Width, mainH, false, m.Playlists, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModePlaylistSelectMove:
		mainView = views.RenderPlaylistSelectDialog(m.Width, mainH, true, m.Playlists, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModeQueue:
		mainView = views.RenderQueueView(m.Width, mainH, m.PlayQueue, m.QueueIndex, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModeLyrics:
		var newOffset int
		mainView, newOffset = views.RenderFullscreenLyricsView(m.Width, mainH, m.Player, m.LyricsData, m.LyricsScrollOffset, m.LyricsAutoScroll, m.Styles)
		m.LyricsScrollOffset = newOffset

	case components.ViewModeIntro:
		mainView = views.RenderIntroView(m.Width, mainH, m.Styles)
	}

	statusBar := components.RenderStatusBar(m.Width, m.Player, m.Styles)
	helpBar := components.RenderHelpBar(m.Width, m.Mode, m.StatusMessage, m.Autoplay, m.Styles)

	return mainView + "\n" + statusBar + "\n" + helpBar
}
