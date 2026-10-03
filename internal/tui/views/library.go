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

// LibraryViewState encapsulates directory browser navigation state.
type LibraryViewState struct {
	CurrentPath   string
	Items         []library.LibraryItem
	SelectedIndex int
	ScrollOffset  int
	FilterQuery   string
}

// RenderLibraryState renders the library view using a structured LibraryViewState.
func RenderLibraryState(width, height int, state LibraryViewState, styles theme.Styles) string {
	return RenderLibraryFilteredView(width, height, state.CurrentPath, state.Items, state.SelectedIndex, state.ScrollOffset, state.FilterQuery, styles)
}

// RenderLibraryView renders the local music directory browser without filter.
func RenderLibraryView(width, height int, currentPath string, items []library.LibraryItem, selectedIndex, scrollOffset int, styles theme.Styles) string {
	return RenderLibraryFilteredView(width, height, currentPath, items, selectedIndex, scrollOffset, "", styles)
}

// RenderLibraryFilteredView renders the local music directory browser with an optional filter query.
func RenderLibraryFilteredView(width, height int, currentPath string, items []library.LibraryItem, selectedIndex, scrollOffset int, filterQuery string, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2
	if innerW < 10 {
		innerW = 10
	}

	lines := make([]string, innerH)
	if len(items) == 0 {
		msg := "Directory is empty."
		if filterQuery != "" {
			msg = "No files matching filter: " + filterQuery
		}
		pad := (innerW - lipgloss.Width(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		for i := 0; i < innerH; i++ {
			if i == innerH/2 {
				lines[i] = stringutil.SafeRepeat(" ", pad) + styles.StatusDim.Render(msg)
			} else {
				lines[i] = strings.Repeat(" ", innerW)
			}
		}
	} else {
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
	}

	title := fmt.Sprintf("LIBRARY: %s", currentPath)
	if filterQuery != "" {
		title = fmt.Sprintf("LIBRARY: %s [/ %s]", currentPath, filterQuery)
	}
	return components.RenderBoxWithTitle(title, lines, width, height, styles)
}
