package views

import (
	"fmt"
	"strings"
	"vibe-fi/internal/config"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
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
		pad := (innerW - len(logoLine)) / 2
		if pad < 0 {
			pad = 0
		}
		if startY+i < innerH {
			lines[startY+i] = strings.Repeat(" ", pad) + styles.StatusTitle.Render(logoLine)
		}
	}

	welcome := fmt.Sprintf("Vibe-Fi Terminal Music Player (v%s)", config.Version)
	welcomePad := (innerW - len(welcome)) / 2
	if welcomePad < 0 {
		welcomePad = 0
	}
	welcomeY := startY + len(asciiLogo) + 2
	if welcomeY < innerH {
		lines[welcomeY] = strings.Repeat(" ", welcomePad) + styles.StatusTitle.Render(welcome)
	}

	instruction := "[L] Library   [S] Search   [P] Playlists   [R] Resume   [ESC] Quit"
	instPad := (innerW - len(instruction)) / 2
	if instPad < 0 {
		instPad = 0
	}
	instY := welcomeY + 2
	if instY < innerH {
		lines[instY] = strings.Repeat(" ", instPad) + styles.HelpKey.Render(instruction)
	}

	return components.RenderBoxWithTitle("", lines, width, height, styles)
}
