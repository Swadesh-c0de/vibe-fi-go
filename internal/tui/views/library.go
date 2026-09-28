package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
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

		if lipgloss.Width(dispName) > innerW-2 {
			dispName = ansi.Truncate(dispName, innerW-5, "...")
		}

		line := " " + dispName
		pad := innerW - lipgloss.Width(line)
		fullLine := line + stringutil.SafeRepeat(" ", pad)

		if idx == selectedIndex {
			lines[i] = styles.SelectedRow.Render(fullLine)
		} else {
			lines[i] = fullLine
		}
	}

	title := fmt.Sprintf("LIBRARY: %s", currentPath)
	return components.RenderBoxWithTitle(title, lines, width, height, styles)
}
