package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// RenderStatusBar renders the 5-line status bar at the bottom.
func RenderStatusBar(width int, p player.AudioPlayer, styles theme.Styles) string {
	statusTag := "NOW PLAYING"
	if p.IsLoading() {
		statusTag = "NOW PLAYING [FETCHING]"
	} else if p.IsBuffering() {
		statusTag = "NOW PLAYING [BUFFERING]"
	} else if p.IsPaused() {
		statusTag = "NOW PLAYING [PAUSED]"
	} else if p.IsIdle() {
		statusTag = "NOW PLAYING [IDLE]"
	}

	innerW := width - 2
	if innerW < 10 {
		innerW = 10
	}

	// Line 1: Centered Title
	title := p.GetMetadata("media-title")
	if title == "" {
		title = p.GetMetadata("filename")
		if title == "" {
			if p.IsLoading() {
				title = "Fetching audio stream..."
			} else if p.IsBuffering() {
				title = "Buffering audio stream..."
			} else {
				if p.IsPlaying() {
					title = "Playing Audio Stream"
				} else {
					title = "Not Playing"
				}
			}
		}
	}

	if lipgloss.Width(title) > innerW-2 {
		title = ansi.Truncate(title, innerW-5, "...")
	}
	tWidth := lipgloss.Width(title)
	titleLeft := (innerW - tWidth) / 2
	titleRight := innerW - titleLeft - tWidth
	line1 := stringutil.SafeRepeat(" ", titleLeft) + styles.StatusTitle.Render(title) + stringutil.SafeRepeat(" ", titleRight)

	// Line 2: Progress Bar
	pos := p.Position()
	dur := p.Duration()
	barWidth := innerW - 4

	var line2 string
	if dur > 0 && barWidth > 4 {
		filled := int((pos / dur) * float64(barWidth-2))
		if filled > barWidth-2 {
			filled = barWidth - 2
		}
		if filled < 0 {
			filled = 0
		}
		empty := (barWidth - 2) - filled

		barContent := "[" + styles.ProgressBar.Render(strings.Repeat("=", filled)) + stringutil.SafeRepeat(" ", empty) + "]"
		barPad := (innerW - barWidth) / 2
		rightPad := innerW - barPad - barWidth
		line2 = stringutil.SafeRepeat(" ", barPad) + barContent + stringutil.SafeRepeat(" ", rightPad)
	} else if p.IsLoading() {
		hint := "Connecting to audio stream..."
		hWidth := lipgloss.Width(hint)
		hPad := (innerW - hWidth) / 2
		rPad := innerW - hPad - hWidth
		line2 = stringutil.SafeRepeat(" ", hPad) + styles.ProgressBar.Render(hint) + stringutil.SafeRepeat(" ", rPad)
	} else if p.IsBuffering() {
		hint := "Buffering audio cache..."
		hWidth := lipgloss.Width(hint)
		hPad := (innerW - hWidth) / 2
		rPad := innerW - hPad - hWidth
		line2 = stringutil.SafeRepeat(" ", hPad) + styles.ProgressBar.Render(hint) + stringutil.SafeRepeat(" ", rPad)
	} else {
		line2 = stringutil.SafeRepeat(" ", innerW)
	}

	// Line 3: Duration on left, Volume on right
	timeStr := stringutil.FormatDuration(pos) + " / " + stringutil.FormatDuration(dur)
	volStr := fmt.Sprintf("Vol: %d%%", p.Volume())

	spaceCount := innerW - 2 - lipgloss.Width(timeStr) - lipgloss.Width(volStr)
	if spaceCount < 1 {
		spaceCount = 1
	}
	line3 := " " + timeStr + stringutil.SafeRepeat(" ", spaceCount) + volStr + " "

	innerLines := []string{line1, line2, line3}
	return RenderBoxWithTitle(statusTag, innerLines, width, 5, styles)
}
