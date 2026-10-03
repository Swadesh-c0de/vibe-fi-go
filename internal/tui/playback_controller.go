package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"vibe-fi/internal/eventbus"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/tui/views"
	"vibe-fi/internal/utils/net"
	"vibe-fi/internal/utils/stringutil"
)

// StartTrackPlayback loads and plays a track, publishing domain events and requesting lyrics.
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

	durSec := stringutil.ParseDuration(duration)

	if m.EventBus != nil {
		m.EventBus.Publish(eventbus.EventTrackStarted, eventbus.TrackStartedEvent{
			Title:    title,
			Artist:   artistHint,
			URL:      url,
			Duration: durSec,
			Playlist: m.PlayingPlaylistName,
		})
		m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{
			State: eventbus.StatePlaying,
		})
	}

	m.saveCurrentState()
	return m.fetchLyricsCmd(title, url, durSec, artistHint)
}

func (m *AppModel) fetchLyricsCmd(title, filePath string, duration float64, artistHint string) tea.Cmd {
	m.LyricsRequestID++
	reqID := m.LyricsRequestID
	mgr := m.LyricsManager

	return func() tea.Msg {
		artist, track := stringutil.CleanTrackTitle(title)
		if artistHint != "" && artist == "" {
			artist = artistHint
		}
		data, err := mgr.FetchLyricsWithFile(filePath, artist, track, duration)
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

func (m *AppModel) playNext() tea.Cmd {
	if len(m.PlayQueue) == 0 {
		return nil
	}
	nextIdx := m.QueueIndex + 1
	if nextIdx >= len(m.PlayQueue) {
		if m.Autoplay {
			nextIdx = 0
		} else {
			return nil
		}
	}
	m.QueueIndex = nextIdx
	song := m.PlayQueue[m.QueueIndex]
	if stringutil.IsURL(song.URL) && !net.IsOnline() {
		return m.ShowStatus("Network unavailable: cannot play next track.")
	}

	statusCmd := m.ShowStatus("Playing: " + song.Title)
	playCmd := m.StartTrackPlayback(song.Title, song.URL, song.Duration, m.PlayingPlaylistName)
	return tea.Batch(statusCmd, playCmd)
}

func (m *AppModel) playPrevious() tea.Cmd {
	if len(m.PlayQueue) == 0 {
		return nil
	}
	prevIdx := m.QueueIndex - 1
	if prevIdx < 0 {
		if m.Autoplay {
			prevIdx = len(m.PlayQueue) - 1
		} else {
			prevIdx = 0
		}
	}
	m.QueueIndex = prevIdx
	song := m.PlayQueue[m.QueueIndex]
	if stringutil.IsURL(song.URL) && !net.IsOnline() {
		return m.ShowStatus("Network unavailable: cannot play previous track.")
	}

	statusCmd := m.ShowStatus("Playing: " + song.Title)
	playCmd := m.StartTrackPlayback(song.Title, song.URL, song.Duration, m.PlayingPlaylistName)
	return tea.Batch(statusCmd, playCmd)
}

// handlePlaybackKey processes keyboard shortcuts when in playback view.
func (m *AppModel) handlePlaybackKey(keyStr string) tea.Cmd {
	switch keyStr {
	case "esc", "q", "Q":
		m.ShowConfirmQuit = true
		m.ConfirmQuitSelection = 1
		return nil

	case " ":
		_ = m.Player.TogglePause()
		if m.EventBus != nil {
			st := eventbus.StatePlaying
			if m.Player.IsPaused() {
				st = eventbus.StatePaused
			}
			m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: st})
		}
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
			if m.EventBus != nil {
				m.EventBus.Publish(eventbus.EventPlaybackStateChanged, eventbus.PlaybackStateChangedEvent{State: eventbus.StatePlaying})
			}
			m.LyricsScrollOffset = 0
			m.LyricsAutoScroll = true
			statusCmd := m.ShowStatus("Replaying...")
			durSec := stringutil.ParseDuration("")
			fetchCmd := m.fetchLyricsCmd(m.CurrentLyricsTitle, m.LastPlayedPath, durSec, "")
			return tea.Batch(statusCmd, fetchCmd)
		}
		return nil

	case "left":
		_ = m.Player.Seek(-5.0)
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventSeeked, eventbus.SeekedEvent{PositionSeconds: m.Player.Position()})
		}
		return nil

	case "right":
		_ = m.Player.Seek(5.0)
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventSeeked, eventbus.SeekedEvent{PositionSeconds: m.Player.Position()})
		}
		return nil

	case "+", "=":
		newVol := m.Player.Volume() + 5
		_ = m.Player.SetVolume(newVol)
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventVolumeChanged, eventbus.VolumeChangedEvent{Volume: m.Player.Volume()})
		}
		return nil

	case "-", "_":
		newVol := m.Player.Volume() - 5
		_ = m.Player.SetVolume(newVol)
		if m.EventBus != nil {
			m.EventBus.Publish(eventbus.EventVolumeChanged, eventbus.VolumeChangedEvent{Volume: m.Player.Volume()})
		}
		return nil

	case "a", "A":
		m.LyricsAutoScroll = !m.LyricsAutoScroll
		if m.LyricsAutoScroll {
			return m.ShowStatus("Lyrics Auto-Scroll: ON")
		}
		return m.ShowStatus("Lyrics Auto-Scroll: OFF")

	case "o", "O":
		m.Autoplay = !m.Autoplay
		autoStr := "OFF"
		if m.Autoplay {
			autoStr = "ON"
		}
		return m.ShowStatus("Autoplay: " + autoStr)

	case "v", "V":
		switch m.PlaybackLayout {
		case views.LayoutSplit:
			m.PlaybackLayout = views.LayoutFullVisualizer
			return m.ShowStatus("Layout: Cinema Visualizer")
		case views.LayoutFullVisualizer:
			m.PlaybackLayout = views.LayoutFullLyrics
			return m.ShowStatus("Layout: Fullscreen Lyrics")
		default:
			m.PlaybackLayout = views.LayoutSplit
			return m.ShowStatus("Layout: Split (Visualizer + Lyrics)")
		}

	case "t", "T":
		m.Theme = theme.CycleTheme(m.Theme.Name)
		m.Styles = theme.MakeStyles(m.Theme)
		m.saveCurrentState()
		return m.ShowStatus("Theme: " + m.Theme.Name)

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

	return nil
}
