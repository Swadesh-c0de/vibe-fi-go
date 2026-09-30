package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// RenderBoxWithTitle renders an inner body wrapped in standard box borders with an optional title.
func RenderBoxWithTitle(title string, innerLines []string, width, height int, styles theme.Styles) string {
	if width < 4 || height < 3 {
		return ""
	}

	bStyle := styles.BorderLine
	titleStyle := styles.BorderTitle

	tl := styles.BorderTL
	if tl == "" {
		tl = bStyle.Render("┌")
	}
	tr := styles.BorderTR
	if tr == "" {
		tr = bStyle.Render("┐")
	}
	bl := styles.BorderBL
	if bl == "" {
		bl = bStyle.Render("└")
	}
	br := styles.BorderBR
	if br == "" {
		br = bStyle.Render("┘")
	}
	bv := styles.BorderV
	if bv == "" {
		bv = bStyle.Render("│")
	}
	bh := styles.BorderH
	if bh == "" {
		bh = bStyle.Render("─")
	}

	innerW := width - 2
	innerH := height - 2

	var out strings.Builder

	// 1. Top border: ┌─ TITLE ──────┐
	out.WriteString(tl)
	if title != "" {
		formattedTitle := " " + title + " "
		tWidth := lipgloss.Width(formattedTitle)
		if tWidth > innerW-2 {
			maxTitle := innerW - 5
			if maxTitle < 1 {
				maxTitle = 1
			}
			formattedTitle = " " + ansi.Truncate(title, maxTitle, "...") + " "
			tWidth = lipgloss.Width(formattedTitle)
		}
		out.WriteString(bh)
		out.WriteString(titleStyle.Render(formattedTitle))
		rem := innerW - 1 - tWidth
		if rem < 0 {
			rem = 0
		}
		out.WriteString(bStyle.Render(strings.Repeat("─", rem)))
	} else {
		out.WriteString(bStyle.Render(strings.Repeat("─", innerW)))
	}
	out.WriteString(tr)
	out.WriteString("\n")

	// 2. Middle rows: │ line │
	for i := 0; i < innerH; i++ {
		line := ""
		if i < len(innerLines) {
			line = innerLines[i]
		}

		lineW := lipgloss.Width(line)
		if lineW > innerW {
			line = ansi.Truncate(line, innerW, "")
			lineW = lipgloss.Width(line)
		}
		pad := innerW - lineW
		if pad < 0 {
			pad = 0
		}

		out.WriteString(bv)
		out.WriteString(line)
		out.WriteString(stringutil.SafeRepeat(" ", pad))
		out.WriteString(bv)
		out.WriteString("\n")
	}

	// 3. Bottom border: └────────────┘
	out.WriteString(bl)
	out.WriteString(bStyle.Render(strings.Repeat("─", innerW)))
	out.WriteString(br)

	return out.String()
}
