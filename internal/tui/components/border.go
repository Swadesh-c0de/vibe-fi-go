package components

import (
	"strings"
	"vibe-fi/internal/tui/theme"
)

// RenderBoxWithTitle renders an inner body wrapped in standard box borders with a title on the top border.
func RenderBoxWithTitle(title string, innerLines []string, width, height int, styles theme.Styles) string {
	if width < 4 || height < 3 {
		return ""
	}

	// Top border
	var topBorder strings.Builder
	topBorder.WriteString("┌")

	if title != "" {
		formattedTitle := " " + title + " "
		styledTitle := styles.BorderTitle.Render(formattedTitle)
		titleRuneLen := len([]rune(formattedTitle))

		topBorder.WriteString("─")
		topBorder.WriteString(styledTitle)
		remaining := width - 2 - 1 - titleRuneLen
		if remaining < 0 {
			remaining = 0
		}
		topBorder.WriteString(styles.BorderBox.Render(strings.Repeat("─", remaining)))
	} else {
		topBorder.WriteString(styles.BorderBox.Render(strings.Repeat("─", width-2)))
	}
	topBorder.WriteString("┐")

	// Bottom border
	bottomBorder := "└" + styles.BorderBox.Render(strings.Repeat("─", width-2)) + "┘"

	innerH := height - 2
	innerW := width - 2

	var out strings.Builder
	out.WriteString(styles.BorderBox.Render("┌") + topBorder.String()[len("┌"):len(topBorder.String())-len("┐")] + styles.BorderBox.Render("┐") + "\n")

	for i := 0; i < innerH; i++ {
		line := ""
		if i < len(innerLines) {
			line = innerLines[i]
		}
		// Pad or truncate
		runes := []rune(line)
		if len(runes) > innerW {
			line = string(runes[:innerW])
		} else if len(runes) < innerW {
			line = line + strings.Repeat(" ", innerW-len(runes))
		}

		out.WriteString(styles.BorderBox.Render("│"))
		out.WriteString(line)
		out.WriteString(styles.BorderBox.Render("│") + "\n")
	}

	out.WriteString(styles.BorderBox.Render(bottomBorder))
	return out.String()
}
