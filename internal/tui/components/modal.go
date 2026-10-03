package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// RenderConfirmQuit renders the centered quit confirmation modal.
func RenderConfirmQuit(termW, termH int, selectedIdx int, styles theme.Styles) string {
	winW := 46
	if termW-4 < winW {
		winW = termW - 4
	}
	if winW < 32 {
		winW = 32
	}
	winH := 7

	prompt := "Wanna quit listening?"
	pWidth := lipgloss.Width(prompt)
	pLeft := (winW - 2 - pWidth) / 2
	line1 := stringutil.SafeRepeat(" ", pLeft) + styles.StatusTitle.Render(prompt)

	line2 := "" // spacing

	btnYes := "[ YES ]"
	btnNo := "[  NO  ]"
	if selectedIdx == 0 {
		btnYes = styles.ModalBtnOn.Render("  YES  ")
		btnNo = styles.ModalBtn.Render("[  NO  ]")
	} else {
		btnYes = styles.ModalBtn.Render("[ YES ]")
		btnNo = styles.ModalBtnOn.Render("   NO   ")
	}

	btnSpace := 6
	btnTotalW := 7 + btnSpace + 8
	btnLeft := (winW - 2 - btnTotalW) / 2
	line3 := stringutil.SafeRepeat(" ", btnLeft) + btnYes + stringutil.SafeRepeat(" ", btnSpace) + btnNo

	box := RenderBoxWithTitle("Confirmation", []string{"", line1, line2, line3, ""}, winW, winH, styles)
	return OverlayCenter(box, termW, termH, winW, winH)
}

// RenderInputPrompt renders a centered input dialog box.
func RenderInputPrompt(termW, termH int, promptTitle, text string, styles theme.Styles) string {
	winW := 50
	if termW-4 < winW {
		winW = termW - 4
	}
	if winW < 20 {
		winW = 20
	}
	winH := 5

	innerW := winW - 2
	inputLine := fmt.Sprintf("> %s_", text)
	if lipgloss.Width(inputLine) > innerW-2 {
		inputLine = "> " + ansi.Truncate(text, innerW-5, "") + "_"
	}
	line := " " + styles.SelectedRow.Render(inputLine)

	box := RenderBoxWithTitle(promptTitle, []string{"", line, ""}, winW, winH, styles)
	return OverlayCenter(box, termW, termH, winW, winH)
}

// RenderHelpModal renders the centered keyboard shortcuts cheat-sheet modal.
func RenderHelpModal(termW, termH int, styles theme.Styles) string {
	winW := 72
	if termW-4 < winW {
		winW = termW - 4
	}
	if winW < 40 {
		winW = 40
	}

	winH := 16
	if termH-2 < winH {
		winH = termH - 2
	}
	if winH < 8 {
		winH = 8
	}

	innerW := winW - 2
	innerH := winH - 2

	type shortcutItem struct {
		key  string
		desc string
	}

	playbackItems := []shortcutItem{
		{"Space", "Play / Pause"},
		{"Left/Right", "Seek -/+ 5s"},
		{"+ / -", "Volume Up / Down"},
		{"N / .", "Next Track"},
		{"B / ,", "Prev Track"},
		{"R", "Replay Song"},
		{"V", "Cycle Layout"},
		{"O", "Toggle Autoplay"},
		{"?", "Quick Reference"},
	}

	navItems := []shortcutItem{
		{"L", "Music Library"},
		{"S", "Search YouTube"},
		{"P", "Playlists Browser"},
		{"C", "Play Queue"},
		{"U", "Stream Custom URL"},
		{"Up / Down", "Scroll Lyrics"},
		{"A", "Auto-Scroll Lyrics"},
		{"T", "Cycle Theme"},
		{"Esc / Q", "Quit / Back"},
	}

	// Layout two columns with a subtle vertical divider
	availableW := innerW - 2 - 3 // 2 spaces margin, 3 chars for " │ "
	if availableW < 10 {
		availableW = 10
	}
	col1W := availableW / 2
	col2W := availableW - col1W

	divider := " " + styles.BorderLine.Render("│") + " "

	formatCol := func(key, desc string, colW int, isHeader bool) string {
		if isHeader {
			hText := styles.ProgressBar.Render(key)
			w := stringutil.Width(key)
			if w < colW {
				return hText + strings.Repeat(" ", colW-w)
			}
			return hText
		}

		keyBadge := "[" + key + "]"
		keyW := 13
		if colW < 20 {
			keyW = colW / 2
		}
		kW := stringutil.Width(keyBadge)
		var keyFormatted string
		if kW < keyW {
			keyFormatted = styles.StatusTitle.Render(keyBadge) + strings.Repeat(" ", keyW-kW)
		} else {
			keyFormatted = styles.StatusTitle.Render(keyBadge) + " "
			keyW = kW + 1
		}

		descW := colW - keyW
		if descW < 1 {
			descW = 1
		}
		descTrunc := stringutil.Truncate(desc, descW, "..")
		dW := stringutil.Width(descTrunc)
		descFormatted := styles.HelpKey.Render(descTrunc)
		if dW < descW {
			descFormatted += strings.Repeat(" ", descW-dW)
		}
		return keyFormatted + descFormatted
	}

	var lines []string
	lines = append(lines, stringutil.SafeRepeat(" ", innerW))

	// Header row
	headerLine := " " + formatCol("PLAYBACK", "", col1W, true) + divider + formatCol("NAVIGATION & VIEWS", "", col2W, true) + " "
	lines = append(lines, headerLine)

	// Items
	maxItems := len(playbackItems)
	if len(navItems) > maxItems {
		maxItems = len(navItems)
	}

	for i := 0; i < maxItems; i++ {
		var col1Str, col2Str string
		if i < len(playbackItems) {
			col1Str = formatCol(playbackItems[i].key, playbackItems[i].desc, col1W, false)
		} else {
			col1Str = strings.Repeat(" ", col1W)
		}

		if i < len(navItems) {
			col2Str = formatCol(navItems[i].key, navItems[i].desc, col2W, false)
		} else {
			col2Str = strings.Repeat(" ", col2W)
		}

		line := " " + col1Str + divider + col2Str + " "
		lines = append(lines, line)
	}

	lines = append(lines, stringutil.SafeRepeat(" ", innerW))

	// Footer
	hint := "Press any key to close"
	hintW := stringutil.Width(hint)
	leftPad := (innerW - hintW) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	rightPad := innerW - leftPad - hintW
	if rightPad < 0 {
		rightPad = 0
	}
	footerLine := stringutil.SafeRepeat(" ", leftPad) + styles.StatusDim.Render(hint) + stringutil.SafeRepeat(" ", rightPad)
	lines = append(lines, footerLine)

	// Clamp to innerH
	if len(lines) > innerH {
		lines = lines[:innerH]
	} else {
		for len(lines) < innerH {
			lines = append(lines, stringutil.SafeRepeat(" ", innerW))
		}
	}

	box := RenderBoxWithTitle("KEYBOARD SHORTCUTS", lines, winW, winH, styles)
	return OverlayCenter(box, termW, termH, winW, winH)
}

// OverlayCenter centers a box over an empty terminal canvas.
func OverlayCenter(box string, termW, termH, boxW, boxH int) string {
	boxLines := strings.Split(box, "\n")
	startY := (termH - boxH) / 2
	startX := (termW - boxW) / 2

	var out strings.Builder
	for y := 0; y < termH; y++ {
		if y >= startY && y < startY+len(boxLines) {
			bLine := boxLines[y-startY]
			out.WriteString(stringutil.SafeRepeat(" ", startX))
			out.WriteString(bLine)
		} else {
			out.WriteString(stringutil.SafeRepeat(" ", termW))
		}
		if y < termH-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}
