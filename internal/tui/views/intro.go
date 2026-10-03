package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/config"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

var modernLogo = []string{
	`██╗   ██╗██╗██████╗ ███████╗    ███████╗██╗`,
	`██║   ██║██║██╔══██╗██╔════╝    ██╔════╝██║`,
	`██║   ██║██║██████╔╝█████╗  ──  █████╗  ██║`,
	`╚██╗ ██╔╝██║██╔══██╗██╔══╝      ██╔══╝  ██║`,
	` ╚████╔╝ ██║██████╔╝███████╗    ██║     ██║`,
	`  ╚═══╝  ╚═╝╚═════╝ ╚══════╝    ╚═╝     ╚═╝`,
}

// IntroViewState contains state needed to render the interactive welcome screen.
type IntroViewState struct {
	SelectedIndex int
	ResumeTitle   string
	PlaylistCount int
	Theme         theme.Theme
	StatusMessage string
}

type introMenuItem struct {
	Key   string
	Title string
}

func getIntroMenuItems() []introMenuItem {
	return []introMenuItem{
		{Key: "L", Title: "Music Library"},
		{Key: "S", Title: "Search YouTube"},
		{Key: "P", Title: "Playlists Manager"},
		{Key: "R", Title: "Resume Session"},
		{Key: "?", Title: "Keyboard Shortcuts"},
		{Key: "Q", Title: "Quit Vibe-Fi"},
	}
}

// RenderIntroState renders the modern, interactive welcome screen with responsive logo and quick actions.
func RenderIntroState(width, height int, state IntroViewState, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2
	if innerH < 1 || innerW < 1 {
		return components.RenderBoxWithTitle("", nil, width, height, styles)
	}

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = stringutil.SafeRepeat(" ", innerW)
	}

	items := getIntroMenuItems()
	themeName := state.Theme.Name
	if themeName == "" {
		themeName = "Midnight"
	}

	useCompact := innerH < 18 || innerW < 47

	curY := 0
	if !useCompact {
		totalContentH := len(modernLogo) + 1 + 1 + len(items) + 2 + 1 + 1
		startY := (innerH - totalContentH) / 2
		if startY < 1 {
			startY = 1
		}
		curY = startY

		// 1. Logo with gradient coloring
		for i, logoLine := range modernLogo {
			lWidth := lipgloss.Width(logoLine)
			pad := (innerW - lWidth) / 2
			if pad < 0 {
				pad = 0
			}
			var styledLine string
			switch i {
			case 0, 1:
				if state.Theme.VizHighColor != "" {
					styledLine = lipgloss.NewStyle().Foreground(state.Theme.VizHighColor).Bold(true).Render(logoLine)
				} else {
					styledLine = styles.StatusTitle.Bold(true).Render(logoLine)
				}
			case 2, 3:
				if state.Theme.ProgressColor != "" {
					styledLine = lipgloss.NewStyle().Foreground(state.Theme.ProgressColor).Bold(true).Render(logoLine)
				} else {
					styledLine = styles.StatusTitle.Render(logoLine)
				}
			default:
				if state.Theme.BorderColor != "" {
					styledLine = lipgloss.NewStyle().Foreground(state.Theme.BorderColor).Bold(true).Render(logoLine)
				} else {
					styledLine = styles.BorderLine.Render(logoLine)
				}
			}
			if curY < innerH {
				rPad := innerW - pad - lWidth
				if rPad < 0 {
					rPad = 0
				}
				lines[curY] = stringutil.SafeRepeat(" ", pad) + styledLine + stringutil.SafeRepeat(" ", rPad)
			}
			curY++
		}

		// 2. Subtitle & Version badge
		subText := fmt.Sprintf("v%s • %s Theme", config.Version, themeName)
		subW := lipgloss.Width(subText)
		subPad := (innerW - subW) / 2
		if subPad < 0 {
			subPad = 0
		}
		if curY < innerH {
			rSubPad := innerW - subPad - subW
			if rSubPad < 0 {
				rSubPad = 0
			}
			lines[curY] = stringutil.SafeRepeat(" ", subPad) + styles.StatusDim.Render(subText) + stringutil.SafeRepeat(" ", rSubPad)
		}
		curY += 2
	} else {
		// Compact header for smaller screens
		compactTitle := fmt.Sprintf("── VIBE-FI • v%s • %s ──", config.Version, themeName)
		cW := lipgloss.Width(compactTitle)
		cPad := (innerW - cW) / 2
		if cPad < 0 {
			cPad = 0
		}
		if cPad+cW > innerW {
			compactTitle = ansi.Truncate(compactTitle, innerW, "")
			cPad = 0
			cW = lipgloss.Width(compactTitle)
		}
		lines[0] = stringutil.SafeRepeat(" ", cPad) + styles.StatusTitle.Render(compactTitle) + stringutil.SafeRepeat(" ", innerW-cPad-cW)
		curY = 2
	}

	// 3. Actions Card
	cardW := 34
	if innerW-4 < cardW {
		cardW = innerW - 4
	}
	if cardW < 24 {
		cardW = innerW
	}
	cardLeft := (innerW - cardW) / 2
	if cardLeft < 0 {
		cardLeft = 0
	}

	innerCardW := cardW - 2
	if innerCardW > 0 && curY < innerH {
		// Card top border
		cardTitle := " QUICK ACTIONS "
		dashCount := innerCardW - lipgloss.Width(cardTitle) - 1
		var cardTop string
		if dashCount >= 1 {
			cardTop = styles.BorderLine.Render("┌─") + styles.BorderTitle.Render(cardTitle) + styles.BorderLine.Render(stringutil.SafeRepeat("─", dashCount)) + styles.BorderLine.Render("┐")
		} else {
			cardTop = styles.BorderLine.Render("┌" + stringutil.SafeRepeat("─", innerCardW) + "┐")
		}
		lines[curY] = stringutil.SafeRepeat(" ", cardLeft) + cardTop + stringutil.SafeRepeat(" ", innerW-cardLeft-cardW)
		curY++

		// Card rows
		for i, item := range items {
			if curY >= innerH-1 {
				break
			}
			prefix := "  "
			if i == state.SelectedIndex {
				prefix = "> "
			}

			keyBadge := fmt.Sprintf("[%s]", item.Key)
			keyFormatted := fmt.Sprintf("%-4s", keyBadge)

			rowText := fmt.Sprintf("%s%s  %s", prefix, keyFormatted, item.Title)
			if lipgloss.Width(rowText) > innerCardW-2 {
				rowText = ansi.Truncate(rowText, innerCardW-2, "")
			}
			pad := innerCardW - lipgloss.Width(rowText)
			if pad < 0 {
				pad = 0
			}
			fullRow := " " + rowText + stringutil.SafeRepeat(" ", pad-1)

			var renderedRow string
			if i == state.SelectedIndex {
				renderedRow = styles.SelectedRow.Render(fullRow)
			} else {
				renderedRow = styles.HelpKey.Render(fullRow)
			}

			cardLine := styles.BorderLine.Render("│") + renderedRow + styles.BorderLine.Render("│")
			lines[curY] = stringutil.SafeRepeat(" ", cardLeft) + cardLine + stringutil.SafeRepeat(" ", innerW-cardLeft-cardW)
			curY++
		}

		// Card bottom border
		if curY < innerH {
			cardBottom := styles.BorderLine.Render("└" + stringutil.SafeRepeat("─", innerCardW) + "┘")
			lines[curY] = stringutil.SafeRepeat(" ", cardLeft) + cardBottom + stringutil.SafeRepeat(" ", innerW-cardLeft-cardW)
			curY++
		}
	}

	curY += 2
	if curY < innerH {
		navText := "[↑/↓] Select   [ENTER] Open   [T] Theme   [?] Help   [ESC/Q] Quit"
		nW := lipgloss.Width(navText)
		nPad := (innerW - nW) / 2
		if nPad < 0 {
			nPad = 0
		}
		rNPad := innerW - nPad - nW
		if rNPad < 0 {
			navText = ansi.Truncate(navText, innerW, "")
			nPad = 0
			rNPad = innerW - lipgloss.Width(navText)
		}
		lines[curY] = stringutil.SafeRepeat(" ", nPad) + styles.StatusDim.Render(navText) + stringutil.SafeRepeat(" ", rNPad)
	}

	return components.RenderBoxWithTitle("", lines, width, height, styles)
}

// RenderIntroView renders the welcome screen (backwards-compatible wrapper).
func RenderIntroView(width, height int, styles theme.Styles) string {
	return RenderIntroState(width, height, IntroViewState{}, styles)
}
