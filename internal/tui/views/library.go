package views

import (
	"fmt"
	"strings"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
)

// RenderLibraryView renders the local music directory browser.
func RenderLibraryView(width, height int, currentPath string, items []library.LibraryItem, selectedIndex, scrollOffset int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2
	if innerW < 10 {
		innerW = 10
	}

	lines := make([]string, innerH)
	for i := 0; i < innerH; i++ {
		idx := i + scrollOffset
		if idx >= len(items) {
			lines[i] = strings.Repeat(" ", innerW)
			continue
		}

		item := items[idx]
		prefix := "      "
		if item.IsDirectory {
			prefix = "[DIR] "
		}

		dispName := prefix + item.Name
		if !item.IsDirectory && item.Duration != "" {
			dispName += " (" + item.Duration + ")"
		}

		runes := []rune(dispName)
		if len(runes) > innerW-2 {
			dispName = string(runes[:innerW-5]) + "..."
		}

		line := " " + dispName
		pad := innerW - len([]rune(line))
		if pad < 0 {
			pad = 0
		}
		fullLine := line + strings.Repeat(" ", pad)

		if idx == selectedIndex {
			lines[i] = styles.SelectedRow.Render(fullLine)
		} else {
			lines[i] = fullLine
		}
	}

	title := fmt.Sprintf("LIBRARY: %s", currentPath)
	return components.RenderBoxWithTitle(title, lines, width, height, styles)
}
