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
