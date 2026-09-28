package views

import (
	"fmt"
	"strings"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
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
	pLeft := (innerW - len(prompt)) / 2
	if pLeft < 0 {
		pLeft = 0
	}

	promptY := innerH/2 - 2
	if promptY < 0 {
		promptY = 0
	}
	lines[promptY] = strings.Repeat(" ", pLeft) + styles.StatusTitle.Render(prompt)

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
		if len(dispQuery) > boxWidth-4 {
			dispQuery = dispQuery[len(dispQuery)-(boxWidth-4):]
		}
		cursorChar := "_"
		field := fmt.Sprintf("> %s%s", dispQuery, cursorChar)
		padField := boxWidth - len(field)
		if padField > 0 {
			field += strings.Repeat(" ", padField)
		}
		lines[boxY] = strings.Repeat(" ", boxX) + styles.SelectedRow.Render(field)
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
		pad := (innerW - len(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		lines[innerH/2] = strings.Repeat(" ", pad) + styles.StatusDim.Render(msg)
		return components.RenderBoxWithTitle("SEARCH RESULTS", lines, width, height, styles)
	}

	titleColW := innerW - 20
	if titleColW < 15 {
		titleColW = 15
	}

	// Header row
	header := fmt.Sprintf(" %-4s %-*s %10s", "#", titleColW, "Title", "Duration")
	if len(header) < innerW {
		header += strings.Repeat(" ", innerW-len(header))
	}
	lines[0] = styles.HeaderRow.Render(header)

	visibleRows := innerH - 1
	for i := 0; i < visibleRows; i++ {
		idx := i + scrollOffset
		if idx >= len(results) {
			break
		}

		item := results[idx]
		title := item.Title
		runes := []rune(title)
		if len(runes) > titleColW {
			title = string(runes[:titleColW-3]) + "..."
		}

		row := fmt.Sprintf(" %-4d %-*s %10s", idx+1, titleColW, title, item.Duration)
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

	return components.RenderBoxWithTitle("SEARCH RESULTS", lines, width, height, styles)
}
