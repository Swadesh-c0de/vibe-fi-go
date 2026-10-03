package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// PlaylistBrowserState encapsulates playlist listing and preview state.
type PlaylistBrowserState struct {
	Playlists     []playlist.Playlist
	PreviewSongs  []playlist.PlaylistSong
	SelectedIndex int
	ScrollOffset  int
}

// RenderPlaylistsBrowserState renders the playlists browser using a structured state.
func RenderPlaylistsBrowserState(width, height int, state PlaylistBrowserState, styles theme.Styles) string {
	return RenderPlaylistsBrowser(width, height, state.Playlists, state.PreviewSongs, state.SelectedIndex, state.ScrollOffset, styles)
}

// PlaylistSongsState encapsulates playlist song browsing state.
type PlaylistSongsState struct {
	PlaylistName  string
	Songs         []playlist.PlaylistSong
	SelectedIndex int
	ScrollOffset  int
}

// RenderPlaylistSongsState renders the songs within a playlist using a structured state.
func RenderPlaylistSongsState(width, height int, state PlaylistSongsState, styles theme.Styles) string {
	return RenderPlaylistSongsView(width, height, state.PlaylistName, state.Songs, state.SelectedIndex, state.ScrollOffset, styles)
}

// RenderPlaylistsBrowser renders the split view (playlists list on left, preview on right).
func RenderPlaylistsBrowser(width, height int, playlists []playlist.Playlist, previewSongs []playlist.PlaylistSong, selectedIndex, scrollOffset int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	if len(playlists) == 0 {
		msg := "No playlists found. Press [N] to create one."
		pad := (innerW - lipgloss.Width(msg)) / 2
		lines[innerH/2] = stringutil.SafeRepeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle("PLAYLISTS", lines, width, height, styles)
	}

	leftW := int(float64(innerW) * 0.35)
	if leftW < 25 {
		leftW = 25
	}
	rightW := innerW - leftW - 1
	if rightW < 10 {
		rightW = 10
	}

	// Left header
	leftHeader := fmt.Sprintf(" %-*s", leftW-1, "Playlist Name")
	if lipgloss.Width(leftHeader) > leftW {
		leftHeader = ansi.Truncate(leftHeader, leftW, "")
	} else if lipgloss.Width(leftHeader) < leftW {
		leftHeader += stringutil.SafeRepeat(" ", leftW-lipgloss.Width(leftHeader))
	}

	// Right header
	rightHeader := " Preview"
	if selectedIndex >= 0 && selectedIndex < len(playlists) {
		rightHeader = fmt.Sprintf(" Preview: %s", playlists[selectedIndex].Name)
	}
	if lipgloss.Width(rightHeader) > rightW-2 {
		rightHeader = ansi.Truncate(rightHeader, rightW-3, "...")
	}
	rightHeader += stringutil.SafeRepeat(" ", rightW-lipgloss.Width(rightHeader))

	lines[0] = styles.HeaderRow.Render(leftHeader) + styles.BorderLine.Render("│") + styles.StatusTitle.Render(rightHeader)

	visiblePlaylists := innerH - 1
	for i := 0; i < visiblePlaylists; i++ {
		pIdx := i + scrollOffset
		var leftCol string
		if pIdx < len(playlists) {
			p := playlists[pIdx]
			countStr := fmt.Sprintf(" (%d)", p.SongCount)
			maxNameLen := leftW - 2 - lipgloss.Width(countStr)
			if maxNameLen < 1 {
				maxNameLen = 1
			}
			name := p.Name
			if lipgloss.Width(name) > maxNameLen {
				name = ansi.Truncate(name, maxNameLen, "...")
			}
			rowText := fmt.Sprintf(" %s%s", name, countStr)
			pad := leftW - lipgloss.Width(rowText)
			rowText += stringutil.SafeRepeat(" ", pad)

			if pIdx == selectedIndex {
				leftCol = styles.SelectedRow.Render(rowText)
			} else {
				leftCol = rowText
			}
		} else {
			leftCol = stringutil.SafeRepeat(" ", leftW)
		}

		// Right column (preview songs)
		var rightCol string
		if i == 0 && len(previewSongs) > 0 {
			maxTitleLen := rightW - 20
			if maxTitleLen < 10 {
				maxTitleLen = 10
			}
			headerRow := fmt.Sprintf(" %-4s %-*s %10s", "#", maxTitleLen, "Title", "Duration")
			pad := rightW - lipgloss.Width(headerRow)
			headerRow += stringutil.SafeRepeat(" ", pad)
			rightCol = styles.HeaderRow.Render(headerRow)
		} else if i > 0 && (i-1) < len(previewSongs) {
			s := previewSongs[i-1]
			maxTitleLen := rightW - 20
			if maxTitleLen < 10 {
				maxTitleLen = 10
			}
			title := s.Title
			if lipgloss.Width(title) > maxTitleLen {
				title = ansi.Truncate(title, maxTitleLen, "...")
			}
			row := fmt.Sprintf(" %-4d %-*s %10s", i, maxTitleLen, title, s.Duration)
			pad := rightW - lipgloss.Width(row)
			row += stringutil.SafeRepeat(" ", pad)
			rightCol = row
		} else if len(previewSongs) == 0 && i == 1 {
			msg := " Playlist is empty."
			rightCol = styles.StatusDim.Render(msg) + stringutil.SafeRepeat(" ", rightW-lipgloss.Width(msg))
		} else {
			rightCol = stringutil.SafeRepeat(" ", rightW)
		}

		lines[i+1] = leftCol + styles.BorderLine.Render("│") + rightCol
	}

	return components.RenderBoxWithTitle("PLAYLISTS", lines, width, height, styles)
}

// RenderPlaylistSongsView renders all tracks in an open playlist.
func RenderPlaylistSongsView(width, height int, playlistName string, songs []playlist.PlaylistSong, selectedIndex, scrollOffset int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	if len(songs) == 0 {
		msg := "Playlist is empty."
		pad := (innerW - lipgloss.Width(msg)) / 2
		lines[innerH/2] = stringutil.SafeRepeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle("PLAYLIST: "+playlistName, lines, width, height, styles)
	}

	maxTitleLen := innerW - 20
	if maxTitleLen < 10 {
		maxTitleLen = 10
	}

	header := fmt.Sprintf(" %-4s %-*s %10s", "#", maxTitleLen, "Title", "Duration")
	pad := innerW - lipgloss.Width(header)
	header += stringutil.SafeRepeat(" ", pad)
	lines[0] = styles.HeaderRow.Render(header)

	visibleRows := innerH - 1
	for i := 0; i < visibleRows; i++ {
		idx := i + scrollOffset
		if idx >= len(songs) {
			break
		}
		s := songs[idx]
		title := s.Title
		if lipgloss.Width(title) > maxTitleLen {
			title = ansi.Truncate(title, maxTitleLen, "...")
		}

		row := fmt.Sprintf(" %-4d %-*s %10s", idx+1, maxTitleLen, title, s.Duration)
		rPad := innerW - lipgloss.Width(row)
		row += stringutil.SafeRepeat(" ", rPad)

		if idx == selectedIndex {
			lines[i+1] = styles.SelectedRow.Render(row)
		} else {
			lines[i+1] = row
		}
	}

	return components.RenderBoxWithTitle("PLAYLIST: "+playlistName, lines, width, height, styles)
}

// RenderPlaylistSelectDialog renders target playlist selection for adding or moving songs.
func RenderPlaylistSelectDialog(width, height int, isMove bool, playlists []playlist.Playlist, selectedIndex, scrollOffset int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	title := "SELECT PLAYLIST TO ADD TO"
	if isMove {
		title = "MOVE SONG TO..."
	}

	if len(playlists) == 0 {
		msg := "No playlists found. Press [N] to create one."
		pad := (innerW - lipgloss.Width(msg)) / 2
		lines[innerH/2] = stringutil.SafeRepeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle(title, lines, width, height, styles)
	}

	header := fmt.Sprintf(" %-20s %10s", "Playlist Name", "Songs")
	pad := innerW - lipgloss.Width(header)
	header += stringutil.SafeRepeat(" ", pad)
	lines[0] = styles.HeaderRow.Render(header)

	visibleRows := innerH - 1
	for i := 0; i < visibleRows; i++ {
		idx := i + scrollOffset
		if idx >= len(playlists) {
			break
		}
		p := playlists[idx]
		name := p.Name
		if lipgloss.Width(name) > 20 {
			name = ansi.Truncate(name, 20, "...")
		}
		row := fmt.Sprintf(" %-20s %10d", name, p.SongCount)
		rPad := innerW - lipgloss.Width(row)
		row += stringutil.SafeRepeat(" ", rPad)

		if idx == selectedIndex {
			lines[i+1] = styles.SelectedRow.Render(row)
		} else {
			lines[i+1] = row
		}
	}

	return components.RenderBoxWithTitle(title, lines, width, height, styles)
}
