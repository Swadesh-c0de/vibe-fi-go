package views

import (
	"fmt"
	"strings"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
)

// RenderQueueView renders the upcoming play queue.
func RenderQueueView(width, height int, queue []playlist.PlaylistSong, queueIndex, selectedIndex, scrollOffset int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	if len(queue) == 0 {
		msg := "Queue is empty."
		pad := (innerW - len(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		lines[innerH/2] = strings.Repeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle("PLAY QUEUE", lines, width, height, styles)
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
		runes := []rune(row)
		if len(runes) > innerW-2 {
			row = string(runes[:innerW-5]) + "..."
		}
		pad := innerW - len([]rune(row))
		if pad > 0 {
			row += strings.Repeat(" ", pad)
		}

		if idx == selectedIndex {
			lines[i] = styles.SelectedRow.Render(row)
		} else if idx == queueIndex {
			lines[i] = styles.ActiveSong.Render(row)
		} else {
			lines[i] = row
		}
	}

	return components.RenderBoxWithTitle("PLAY QUEUE", lines, width, height, styles)
}
