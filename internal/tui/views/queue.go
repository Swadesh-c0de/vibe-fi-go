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

// QueueViewState encapsulates play queue display state.
type QueueViewState struct {
	Queue         []playlist.PlaylistSong
	QueueIndex    int
	SelectedIndex int
	ScrollOffset  int
	FilterQuery   string
}

// RenderQueueState renders the queue view using a structured state.
func RenderQueueState(width, height int, state QueueViewState, styles theme.Styles) string {
	return RenderQueueFilteredView(width, height, state.Queue, state.QueueIndex, state.SelectedIndex, state.ScrollOffset, state.FilterQuery, styles)
}

// RenderQueueView renders the upcoming play queue without filter.
func RenderQueueView(width, height int, queue []playlist.PlaylistSong, queueIndex, selectedIndex, scrollOffset int, styles theme.Styles) string {
	return RenderQueueFilteredView(width, height, queue, queueIndex, selectedIndex, scrollOffset, "", styles)
}

// RenderQueueFilteredView renders the upcoming play queue with an optional filter query.
func RenderQueueFilteredView(width, height int, queue []playlist.PlaylistSong, queueIndex, selectedIndex, scrollOffset int, filterQuery string, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	title := "PLAY QUEUE"
	if filterQuery != "" {
		title = fmt.Sprintf("PLAY QUEUE [/ %s]", filterQuery)
	}

	if len(queue) == 0 {
		msg := "Queue is empty."
		if filterQuery != "" {
			msg = "No tracks matching filter: " + filterQuery
		}
		pad := (innerW - lipgloss.Width(msg)) / 2
		lines[innerH/2] = stringutil.SafeRepeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle(title, lines, width, height, styles)
	}

	for i := 0; i < innerH; i++ {
		idx := i + scrollOffset
		if idx >= len(queue) {
			break
		}

		song := queue[idx]
		prefix := "  "
		if idx == queueIndex {
			prefix = "> "
		}

		row := fmt.Sprintf("%s%s (%s)", prefix, song.Title, song.Duration)
		if lipgloss.Width(row) > innerW-2 {
			row = ansi.Truncate(row, innerW-5, "...")
		}
		pad := innerW - lipgloss.Width(row)
		row += stringutil.SafeRepeat(" ", pad)

		switch idx {
		case selectedIndex:
			lines[i] = styles.SelectedRow.Render(row)
		case queueIndex:
			lines[i] = styles.ActiveSong.Render(row)
		default:
			lines[i] = row
		}
	}

	return components.RenderBoxWithTitle(title, lines, width, height, styles)
}
