package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// ViewMode mirrors AppMode from C++.
type ViewMode int

const (
	ViewModePlayback ViewMode = iota
	ViewModeLibrary
	ViewModeSearchInput
	ViewModeSearchResults
	ViewModePlaylistBrowser
	ViewModePlaylistView
	ViewModePlaylistSelectAdd
	ViewModePlaylistSelectMove
	ViewModeQueue
	ViewModeLyrics
	ViewModeIntro
)

// RenderHelpBar renders the 3-line footer help box.
func RenderHelpBar(width int, mode ViewMode, msg string, autoplay bool, styles theme.Styles) string {
	innerW := width - 2
	if innerW < 10 {
		innerW = 10
	}

	var content string
	if msg != "" {
		truncatedMsg := msg
		if lipgloss.Width(truncatedMsg) > innerW-7 {
			maxMsg := innerW - 10
			if maxMsg < 1 {
				maxMsg = 1
			}
			truncatedMsg = ansi.Truncate(truncatedMsg, maxMsg, "...")
		}
		content = styles.HelpAlert.Render("MSG: " + truncatedMsg)
	} else {
		var text string
		switch mode {
		case ViewModePlayback:
			autoStr := "OFF"
			if autoplay {
				autoStr = "ON"
			}
			text = fmt.Sprintf("[SPACE] Pause [A] Add Playlist [N/B] Next/Prev [V] Layout [C] Queue [L] Library [S] Search [P] Playlist [R] Replay [O] Autoplay:%s [?] Help [ESC/Q] Quit", autoStr)
		case ViewModeLibrary:
			text = "[ENTER] Select/Play [/] Filter [BKSP/h] Up [?] Help [ESC] Back"
		case ViewModeSearchInput:
			text = "[ENTER] Search YouTube [ESC] Cancel"
		case ViewModeSearchResults:
			text = "[ENTER] Play [A] Add to Playlist [S] New Search [?] Help [ESC] Back"
		case ViewModePlaylistBrowser:
			text = "[ENTER] View [/] Filter [N] New [D] Delete [R] Rename [E] Export [?] Help [ESC] Back"
		case ViewModePlaylistView:
			text = "[ENTER] Play [/] Filter [D] Remove [M] Move [?] Help [ESC] Back"
		case ViewModePlaylistSelectAdd, ViewModePlaylistSelectMove:
			text = "[ENTER] Select [N] New Playlist [ESC] Cancel"
		case ViewModeQueue:
			text = "[ENTER] Play Selected [A] Add to Playlist [/] Filter [D] Remove [?] Help [ESC] Back"
		case ViewModeLyrics:
			text = "[Y] Auto-Scroll [UP/DOWN] Scroll [?] Help [ESC] Back"
		case ViewModeIntro:
			text = "[↑/↓] Select [ENTER] Open [T] Theme [?] Help [ESC/Q] Quit"
		}

		if lipgloss.Width(text) > innerW-2 {
			maxT := innerW - 5
			if maxT < 1 {
				maxT = 1
			}
			text = ansi.Truncate(text, maxT, "...")
		}
		content = styles.HelpKey.Render(text)
	}

	cWidth := lipgloss.Width(content)
	pad := innerW - 1 - cWidth
	line := " " + content + stringutil.SafeRepeat(" ", pad)

	innerLines := []string{line}
	return RenderBoxWithTitle("", innerLines, width, 3, styles)
}
