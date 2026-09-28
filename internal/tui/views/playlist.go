package views

import (
	"fmt"
	"strings"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
)

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
		pad := (innerW - len(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		lines[innerH/2] = strings.Repeat(" ", pad) + styles.StatusDim.Render(msg)
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
	// Right header
	rightHeader := " Preview"
	if selectedIndex >= 0 && selectedIndex < len(playlists) {
		rightHeader = fmt.Sprintf(" Preview: %s", playlists[selectedIndex].Name)
	}
	if len(rightHeader) > rightW-2 {
		rightHeader = rightHeader[:rightW-5] + "..."
	}
	rightHeader += strings.Repeat(" ", rightW-len(rightHeader))

	lines[0] = styles.HeaderRow.Render(leftHeader) + styles.BorderBox.Render("│") + styles.StatusTitle.Render(rightHeader)

	visiblePlaylists := innerH - 1
	for i := 0; i < visiblePlaylists; i++ {
		pIdx := i + scrollOffset
		var leftCol string
		if pIdx < len(playlists) {
			p := playlists[pIdx]
			countStr := fmt.Sprintf(" (%d)", p.SongCount)
			maxNameLen := leftW - 2 - len(countStr)
			name := p.Name
			if len(name) > maxNameLen && maxNameLen > 3 {
				name = name[:maxNameLen-3] + "..."
			}
			rowText := fmt.Sprintf(" %s%s", name, countStr)
			pad := leftW - len([]rune(rowText))
			if pad > 0 {
				rowText += strings.Repeat(" ", pad)
			}
			if pIdx == selectedIndex {
				leftCol = styles.SelectedRow.Render(rowText)
			} else {
				leftCol = rowText
			}
		} else {
			leftCol = strings.Repeat(" ", leftW)
		}

		// Right column (preview songs)
		var rightCol string
		if i == 0 && len(previewSongs) > 0 {
			maxTitleLen := rightW - 20
			if maxTitleLen < 10 {
				maxTitleLen = 10
			}
			rightCol = styles.HeaderRow.Render(fmt.Sprintf(" %-4s %-*s %10s", "#", maxTitleLen, "Title", "Duration"))
			pad := rightW - len([]rune(rightCol))
			if pad > 0 {
				rightCol += strings.Repeat(" ", pad)
			}
		} else if i > 0 && (i-1) < len(previewSongs) {
			s := previewSongs[i-1]
			maxTitleLen := rightW - 20
			if maxTitleLen < 10 {
				maxTitleLen = 10
			}
			title := s.Title
			runes := []rune(title)
			if len(runes) > maxTitleLen {
				title = string(runes[:maxTitleLen-3]) + "..."
			}
			rightCol = fmt.Sprintf(" %-4d %-*s %10s", i, maxTitleLen, title, s.Duration)
			pad := rightW - len([]rune(rightCol))
			if pad > 0 {
				rightCol += strings.Repeat(" ", pad)
			}
		} else if len(previewSongs) == 0 && i == 1 {
			msg := " Playlist is empty."
			rightCol = styles.StatusDim.Render(msg) + strings.Repeat(" ", rightW-len(msg))
		} else {
			rightCol = strings.Repeat(" ", rightW)
		}

		lines[i+1] = leftCol + styles.BorderBox.Render("│") + rightCol
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
		pad := (innerW - len(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		lines[innerH/2] = strings.Repeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle("PLAYLIST: "+playlistName, lines, width, height, styles)
	}

	maxTitleLen := innerW - 20
	if maxTitleLen < 10 {
		maxTitleLen = 10
	}

	header := fmt.Sprintf(" %-4s %-*s %10s", "#", maxTitleLen, "Title", "Duration")
	if len(header) < innerW {
		header += strings.Repeat(" ", innerW-len(header))
	}
	lines[0] = styles.HeaderRow.Render(header)

	visibleRows := innerH - 1
	for i := 0; i < visibleRows; i++ {
		idx := i + scrollOffset
		if idx >= len(songs) {
			break
		}
		s := songs[idx]
		title := s.Title
		runes := []rune(title)
		if len(runes) > maxTitleLen {
			title = string(runes[:maxTitleLen-3]) + "..."
		}

		row := fmt.Sprintf(" %-4d %-*s %10s", idx+1, maxTitleLen, title, s.Duration)
		pad := innerW - len([]rune(row))
		if pad > 0 {
			row += strings.Repeat(" ", pad)
		}

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
		pad := (innerW - len(msg)) / 2
		lines[innerH/2] = strings.Repeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle(title, lines, width, height, styles)
	}

	header := fmt.Sprintf(" %-20s %10s", "Playlist Name", "Songs")
	if len(header) < innerW {
		header += strings.Repeat(" ", innerW-len(header))
	}
	lines[0] = styles.HeaderRow.Render(header)

	visibleRows := innerH - 1
	for i := 0; i < visibleRows; i++ {
		idx := i + scrollOffset
		if idx >= len(playlists) {
			break
		}
		p := playlists[idx]
		name := p.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		row := fmt.Sprintf(" %-20s %10d", name, p.SongCount)
		pad := innerW - len([]rune(row))
		if pad > 0 {
			row += strings.Repeat(" ", pad)
		}

		if idx == selectedIndex {
			lines[i+1] = styles.SelectedRow.Render(row)
		} else {
			lines[i+1] = row
		}
	}

	return components.RenderBoxWithTitle(title, lines, width, height, styles)
}
