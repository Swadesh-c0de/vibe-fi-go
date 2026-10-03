package tui

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/utils/net"
	"vibe-fi/internal/utils/stringutil"
)

// handleKey routes key events to view-specific handlers.
func (m *AppModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	if msg.Type == tea.KeyCtrlC {
		m.saveCurrentState()
		_ = m.Player.Stop()
		m.Quitting = true
		return tea.Quit
	}

	keyStr := msg.String()

	// Global help cheat-sheet (? or F1) across any non-text-input screen
	if m.Mode != components.ViewModeSearchInput && (keyStr == "?" || keyStr == "f1") {
		m.ShowHelpModal = true
		return nil
	}

	// Playback Mode Hotkeys
	if m.Mode == components.ViewModePlayback {
		return m.handlePlaybackKey(keyStr)
	}

	// Mode-specific input handling
	switch m.Mode {
	case components.ViewModeIntro:
		return m.handleIntroKey(keyStr)

	case components.ViewModeLibrary:
		return m.handleLibraryKey(keyStr)

	case components.ViewModeSearchInput:
		return m.handleSearchInputKey(msg)

	case components.ViewModeSearchResults:
		return m.handleSearchResultsKey(keyStr)

	case components.ViewModePlaylistBrowser:
		return m.handlePlaylistBrowserKey(keyStr)

	case components.ViewModePlaylistView:
		return m.handlePlaylistSongsKey(keyStr)

	case components.ViewModePlaylistSelectAdd, components.ViewModePlaylistSelectMove:
		return m.handlePlaylistSelectKey(keyStr)

	case components.ViewModeQueue:
		return m.handleQueueKey(keyStr)

	case components.ViewModeLyrics:
		return m.handleLyricsKey(keyStr)
	}

	return nil
}

func (m *AppModel) handleIntroKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handleLibraryKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handleSearchInputKey(msg tea.KeyMsg) tea.Cmd {
	keyStr := msg.String()
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
	return nil
}

func (m *AppModel) handleSearchResultsKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handlePlaylistBrowserKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handlePlaylistSongsKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handlePlaylistSelectKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handleQueueKey(keyStr string) tea.Cmd {
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
	return nil
}

func (m *AppModel) handleLyricsKey(keyStr string) tea.Cmd {
	switch keyStr {
	case "esc", "q", "Q":
		m.SetMode(components.ViewModePlayback)
	case "a", "A":
		m.LyricsAutoScroll = !m.LyricsAutoScroll
		if m.LyricsAutoScroll {
			return m.ShowStatus("Lyrics Auto-Scroll: ON")
		}
		return m.ShowStatus("Lyrics Auto-Scroll: OFF")
	case "up", "k":
		if m.LyricsScrollOffset > 0 {
			m.LyricsScrollOffset--
		}
		m.LyricsAutoScroll = false
	case "down", "j":
		m.LyricsScrollOffset++
		m.LyricsAutoScroll = false
	}
	return nil
}

func (m *AppModel) updatePreviewSongs() {
	if len(m.Playlists) > 0 && m.SelectionIndex >= 0 && m.SelectionIndex < len(m.Playlists) {
		name := m.Playlists[m.SelectionIndex].Name
		m.PreviewSongs = m.PlaylistManager.GetPlaylistSongs(name)
	} else {
		m.PreviewSongs = nil
	}
}
