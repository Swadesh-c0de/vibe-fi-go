package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"vibe-fi/internal/config"
	"vibe-fi/internal/eventbus"
	"vibe-fi/internal/integration/discord"
	"vibe-fi/internal/integration/mpris"
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

// AppModel is the central Bubble Tea application state model.
type AppModel struct {
	Player          player.AudioPlayer
	Library         *library.Library
	PlaylistManager *playlist.PlaylistManager
	LyricsManager   *lyrics.LyricsManager
	Visualizer      *visualizer.Visualizer
	MprisServer     *mpris.Server
	DiscordClient   *discord.Client
	EventBus        *eventbus.Bus

	Mode          components.ViewMode
	Width         int
	Height        int
	Theme         theme.Theme
	Styles        theme.Styles
	StatusMessage string
	Autoplay      bool

	trimCounter int

	CurrentPath  string
	LibraryItems []library.LibraryItem

	SearchQuery   string
	SearchResults []search.SearchResult

	Playlists             []playlist.Playlist
	PlaylistSongs         []playlist.PlaylistSong
	CurrentPlaylistName   string
	PreviewSongs          []playlist.PlaylistSong
	IsPlayingFromPlaylist bool
	PlayingPlaylistName   string

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

	// Modals & Overlays
	ShowConfirmQuit      bool
	ConfirmQuitSelection int // 0: YES, 1: NO

	ShowInputPrompt     bool
	InputPromptTitle    string
	InputPromptText     string
	InputPromptCallback func(text string) tea.Cmd

	ShowHelpModal  bool
	PlaybackLayout views.PlaybackLayout

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
		EventBus:             eventbus.New(),
		Mode:                 components.ViewModeIntro,
		Theme:                th,
		Styles:               theme.MakeStyles(th),
		CurrentPath:          homeDir,
		LibraryItems:         items,
		QueueIndex:           -1,
		PlaybackLayout:       views.LayoutSplit,
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

func (m *AppModel) saveCurrentState() {
	st := config.SessionState{
		Path:       m.LastPlayedPath,
		Title:      m.CurrentLyricsTitle,
		Position:   m.Player.Position(),
		Volume:     m.Player.Volume(),
		Index:      m.QueueIndex,
		Theme:      m.Theme.Name,
		Visualizer: 0,
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

		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StatePlaying})
		}

		m.SetMode(components.ViewModePlayback)
		cmd1 := m.ShowStatus("Resuming session...")
		durSec := stringutil.ParseDuration("")
		cmd2 := m.fetchLyricsCmd(st.Title, st.Path, durSec, "")
		return tea.Batch(cmd1, cmd2)
	}

	return m.ShowStatus("No previous track found in session.")
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

		// Natural EOF check for autoplay
		if m.Player.ConsumeTrackFinished() {
			if m.Autoplay {
				nextCmd := m.playNext()
				return m, tea.Batch(nextCmd, m.tickCmd())
			} else {
				if m.EventBus != nil {
					m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StateStopped})
				}
			}
		}

		if m.Player.ConsumePlaybackError() {
			if m.EventBus != nil {
				m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StateStopped})
			}
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

	case UpdateDiscoveredMsg:
		return m, m.ShowStatus(fmt.Sprintf("Update %s available! Run vibe --update to upgrade.", msg.Version))

	case StatusMsg:
		statusCmd := m.ShowStatus(msg.Message)
		playCmd := m.LoadState()
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
		if m.ShowHelpModal {
			m.ShowHelpModal = false
			return m, nil
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
		if m.EventBus != nil {
			st := eventbus.StatePlaying
			if m.Player.IsPaused() {
				st = eventbus.StatePaused
			}
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: st})
		}
	case MprisPlay:
		_ = m.Player.Play()
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StatePlaying})
		}
	case MprisPause:
		_ = m.Player.Pause()
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StatePaused})
		}
	case MprisNext:
		return m.playNext()
	case MprisPrevious:
		return m.playPrevious()
	case MprisStop:
		_ = m.Player.Stop()
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StateStopped})
		}
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
			if m.EventBus != nil {
				m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StateStopped})
			}
			if m.DiscordClient != nil {
				m.DiscordClient.Close()
			}
			m.Quitting = true
			return tea.Quit
		}
		m.ShowConfirmQuit = false
	case "esc", "n", "N":
		m.ShowConfirmQuit = false
	case "y", "Y":
		m.saveCurrentState()
		_ = m.Player.Stop()
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StateStopped})
		}
		if m.DiscordClient != nil {
			m.DiscordClient.Close()
		}
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
	if m.ShowHelpModal {
		return components.RenderHelpModal(m.Width, m.Height, m.Styles)
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
		pbState := views.PlaybackViewState{
			Layout:       m.PlaybackLayout,
			LyricsData:   m.LyricsData,
			ScrollOffset: m.LyricsScrollOffset,
			AutoScroll:   m.LyricsAutoScroll,
		}
		mainView = views.RenderPlaybackState(m.Width, mainH, &pbState, m.Player, m.Visualizer, m.Styles)
		m.LyricsScrollOffset = pbState.ScrollOffset
		m.LyricsAutoScroll = pbState.AutoScroll

	case components.ViewModeLibrary:
		libState := views.LibraryViewState{
			CurrentPath:   m.CurrentPath,
			Items:         m.LibraryItems,
			SelectedIndex: m.SelectionIndex,
			ScrollOffset:  m.ScrollOffset,
		}
		mainView = views.RenderLibraryState(m.Width, mainH, libState, m.Styles)

	case components.ViewModeSearchInput:
		mainView = views.RenderSearchInput(m.Width, mainH, m.SearchQuery, m.Styles)

	case components.ViewModeSearchResults:
		searchState := views.SearchResultsState{
			Results:       m.SearchResults,
			SelectedIndex: m.SelectionIndex,
			ScrollOffset:  m.ScrollOffset,
		}
		mainView = views.RenderSearchResultsState(m.Width, mainH, searchState, m.Styles)

	case components.ViewModePlaylistBrowser:
		browserState := views.PlaylistBrowserState{
			Playlists:     m.Playlists,
			PreviewSongs:  m.PreviewSongs,
			SelectedIndex: m.SelectionIndex,
			ScrollOffset:  m.ScrollOffset,
		}
		mainView = views.RenderPlaylistsBrowserState(m.Width, mainH, browserState, m.Styles)

	case components.ViewModePlaylistView:
		songsState := views.PlaylistSongsState{
			PlaylistName:  m.CurrentPlaylistName,
			Songs:         m.PlaylistSongs,
			SelectedIndex: m.SelectionIndex,
			ScrollOffset:  m.ScrollOffset,
		}
		mainView = views.RenderPlaylistSongsState(m.Width, mainH, songsState, m.Styles)

	case components.ViewModePlaylistSelectAdd:
		mainView = views.RenderPlaylistSelectDialog(m.Width, mainH, false, m.Playlists, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModePlaylistSelectMove:
		mainView = views.RenderPlaylistSelectDialog(m.Width, mainH, true, m.Playlists, m.SelectionIndex, m.ScrollOffset, m.Styles)

	case components.ViewModeQueue:
		queueState := views.QueueViewState{
			Queue:         m.PlayQueue,
			QueueIndex:    m.QueueIndex,
			SelectedIndex: m.SelectionIndex,
			ScrollOffset:  m.ScrollOffset,
		}
		mainView = views.RenderQueueState(m.Width, mainH, queueState, m.Styles)

	case components.ViewModeLyrics:
		var newOffset int
		var newAutoScroll bool
		mainView, newOffset, newAutoScroll = views.RenderFullscreenLyricsView(m.Width, mainH, m.Player, m.LyricsData, m.LyricsScrollOffset, m.LyricsAutoScroll, m.Styles)
		m.LyricsScrollOffset = newOffset
		m.LyricsAutoScroll = newAutoScroll

	case components.ViewModeIntro:
		mainView = views.RenderIntroView(m.Width, mainH, m.Styles)
	}

	statusBar := components.RenderStatusBar(m.Width, m.Player, m.Styles)
	helpBar := components.RenderHelpBar(m.Width, m.Mode, m.StatusMessage, m.Autoplay, m.Styles)

	return mainView + "\n" + statusBar + "\n" + helpBar
}
