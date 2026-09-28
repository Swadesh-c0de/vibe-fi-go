package components

import (
	"fmt"
	"strings"
	"vibe-fi/internal/tui/theme"
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
		content = styles.HelpAlert.Render("MSG: " + msg)
	} else {
		var text string
		switch mode {
		case ViewModePlayback:
			autoStr := "OFF"
			if autoplay {
				autoStr = "ON"
			}
			text = fmt.Sprintf("[SPACE] Pause [N/B] Next/Prev [C] Queue [L] Library [S] Search [P] Playlist [R] Replay [O] Autoplay:%s [ESC/Q] Quit", autoStr)
		case ViewModeLibrary:
			text = "[ENTER] Select/Play [BKSP/h] Parent Directory [ESC] Playback"
		case ViewModeSearchInput:
			text = "[ENTER] Search YouTube [ESC] Cancel"
		case ViewModeSearchResults:
			text = "[ENTER] Play [A] Add to Playlist [S] New Search [ESC] Back"
		case ViewModePlaylistBrowser:
			text = "[ENTER] View [N] New [D] Delete [R] Rename [E] Export M3U [ESC] Back"
		case ViewModePlaylistView:
			text = "[ENTER] Play [D] Remove Song [M] Move Song [ESC] Back"
		case ViewModePlaylistSelectAdd, ViewModePlaylistSelectMove:
			text = "[ENTER] Select [N] New Playlist [ESC] Cancel"
		case ViewModeQueue:
			text = "[ENTER] Play Selected [D] Remove [ESC] Back"
		case ViewModeLyrics:
			text = "[UP/DOWN] Scroll Lyrics [ESC] Back"
		case ViewModeIntro:
			text = "[L] Library [S] Search [P] Playlists [R] Resume [ENTER] Library [ESC] Quit"
		}
		content = styles.HelpKey.Render(text)
	}

	runes := []rune(content)
	if len(runes) > innerW-2 {
		content = string(runes[:innerW-2])
	}
	pad := innerW - len([]rune(content)) - 1
	if pad < 0 {
		pad = 0
	}
	line := " " + content + strings.Repeat(" ", pad)

	innerLines := []string{line}
	return RenderBoxWithTitle("", innerLines, width, 3, styles)
}
