package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// RenderSearchInput renders the search prompt dialog screen.
func RenderSearchInput(width, height int, query string, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	prompt := "What do you want to listen to?"
	pLeft := (innerW - lipgloss.Width(prompt)) / 2
	if pLeft < 0 {
		pLeft = 0
	}

	promptY := innerH/2 - 2
	if promptY < 0 {
		promptY = 0
	}
	if promptY < innerH {
		lines[promptY] = stringutil.SafeRepeat(" ", pLeft) + styles.StatusTitle.Render(prompt)
	}

	boxWidth := 60
	if innerW-6 < boxWidth {
		boxWidth = innerW - 6
	}
	if boxWidth < 10 {
		boxWidth = 10
	}
	boxX := (innerW - boxWidth) / 2
	if boxX < 0 {
		boxX = 0
	}

	boxY := promptY + 2
	if boxY < innerH {
		dispQuery := query
		if lipgloss.Width(dispQuery) > boxWidth-4 {
			dispQuery = ansi.Truncate(dispQuery, boxWidth-4, "")
		}
		cursorChar := "_"
		field := fmt.Sprintf("> %s%s", dispQuery, cursorChar)
		padField := boxWidth - lipgloss.Width(field)
		field += stringutil.SafeRepeat(" ", padField)
		lines[boxY] = stringutil.SafeRepeat(" ", boxX) + styles.SelectedRow.Render(field)
	}

	return components.RenderBoxWithTitle("SEARCH YOUTUBE", lines, width, height, styles)
}

// RenderSearchResults renders the search result table.
func RenderSearchResults(width, height int, results []search.SearchResult, selectedIndex, scrollOffset int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	if len(results) == 0 {
		msg := "No results found or searching..."
		pad := (innerW - lipgloss.Width(msg)) / 2
		lines[innerH/2] = stringutil.SafeRepeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle("SEARCH RESULTS", lines, width, height, styles)
	}

	titleColW := innerW - 20
	if titleColW < 15 {
		titleColW = 15
	}

	// Header row
	header := fmt.Sprintf(" %-4s %-*s %10s", "#", titleColW, "Title", "Duration")
	pad := innerW - lipgloss.Width(header)
	header += stringutil.SafeRepeat(" ", pad)
	lines[0] = styles.HeaderRow.Render(header)

	visibleRows := innerH - 1
	for i := 0; i < visibleRows; i++ {
		idx := i + scrollOffset
		if idx >= len(results) {
			break
		}

		item := results[idx]
		title := item.Title
		if lipgloss.Width(title) > titleColW {
			title = ansi.Truncate(title, titleColW-3, "...")
		}

		row := fmt.Sprintf(" %-4d %-*s %10s", idx+1, titleColW, title, item.Duration)
		rPad := innerW - lipgloss.Width(row)
		row += stringutil.SafeRepeat(" ", rPad)

		if idx == selectedIndex {
			lines[i+1] = styles.SelectedRow.Render(row)
		} else {
			lines[i+1] = row
		}
	}

	return components.RenderBoxWithTitle("SEARCH RESULTS", lines, width, height, styles)
}
