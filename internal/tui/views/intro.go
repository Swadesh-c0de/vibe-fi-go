package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"vibe-fi/internal/config"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

var asciiLogo = []string{
	` __      __  ___   ____    ______           ______   __ `,
	` \ \    / / |_ _| |  _ \  |  ____|         |  ____| |  |`,
	`  \ \  / /   | |  | |_) | | |__     _____  | |__    |  |`,
	`   \ \/ /    | |  |  _ <  |  __|   |_____| |  __|   |  |`,
	`    \  /     | |  | |_) | | |____          | |      |  |`,
	`     \/     |___| |____/  |______|         |_|      |__|`,
}

// RenderIntroView renders the welcome screen with the ASCII art logo and shortcuts.
func RenderIntroView(width, height int, styles theme.Styles) string {
	innerH := height - 2
	innerW := width - 2

	lines := make([]string, innerH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", innerW)
	}

	startY := (innerH-len(asciiLogo))/2 - 3
	if startY < 1 {
		startY = 1
	}

	for i, logoLine := range asciiLogo {
		lWidth := lipgloss.Width(logoLine)
		pad := (innerW - lWidth) / 2
		rPad := innerW - pad - lWidth
		if startY+i < innerH {
			lines[startY+i] = stringutil.SafeRepeat(" ", pad) + styles.StatusTitle.Render(logoLine) + stringutil.SafeRepeat(" ", rPad)
		}
	}

	welcome := fmt.Sprintf("Vibe-Fi Terminal Music Player (v%s)", config.Version)
	wWidth := lipgloss.Width(welcome)
	welcomePad := (innerW - wWidth) / 2
	rWelcomePad := innerW - welcomePad - wWidth
	welcomeY := startY + len(asciiLogo) + 2
	if welcomeY < innerH {
		lines[welcomeY] = stringutil.SafeRepeat(" ", welcomePad) + styles.StatusTitle.Render(welcome) + stringutil.SafeRepeat(" ", rWelcomePad)
	}

	instruction := "[L] Library   [S] Search   [P] Playlists   [R] Resume   [ESC] Quit"
	iWidth := lipgloss.Width(instruction)
	instPad := (innerW - iWidth) / 2
	rInstPad := innerW - instPad - iWidth
	instY := welcomeY + 2
	if instY < innerH {
		lines[instY] = stringutil.SafeRepeat(" ", instPad) + styles.HelpKey.Render(instruction) + stringutil.SafeRepeat(" ", rInstPad)
	}

	return components.RenderBoxWithTitle("", lines, width, height, styles)
}
