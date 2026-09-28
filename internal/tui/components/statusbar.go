package components

import (
	"fmt"
	"strings"
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

	// Line 1: Title
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

	if len(title) > innerW-2 {
		title = title[:innerW-5] + "..."
	}
	titleLeft := (innerW - len(title)) / 2
	if titleLeft < 0 {
		titleLeft = 0
	}
	line1 := strings.Repeat(" ", titleLeft) + styles.StatusTitle.Render(title) + strings.Repeat(" ", innerW-titleLeft-len(title))

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

		barContent := "[" + styles.ProgressBar.Render(strings.Repeat("=", filled)) + strings.Repeat(" ", empty) + "]"
		barPad := (innerW - barWidth) / 2
		if barPad < 0 {
			barPad = 0
		}
		line2 = strings.Repeat(" ", barPad) + barContent + strings.Repeat(" ", innerW-barPad-len([]rune(barContent)))
	} else if p.IsLoading() {
		hint := "Connecting to audio stream..."
		hPad := (innerW - len(hint)) / 2
		line2 = strings.Repeat(" ", hPad) + styles.ProgressBar.Render(hint) + strings.Repeat(" ", innerW-hPad-len(hint))
	} else if p.IsBuffering() {
		hint := "Buffering audio cache..."
		hPad := (innerW - len(hint)) / 2
		line2 = strings.Repeat(" ", hPad) + styles.ProgressBar.Render(hint) + strings.Repeat(" ", innerW-hPad-len(hint))
	} else {
		line2 = strings.Repeat(" ", innerW)
	}

	// Line 3: Duration on left, Volume on right
	timeStr := stringutil.FormatDuration(pos) + " / " + stringutil.FormatDuration(dur)
	volStr := fmt.Sprintf("Vol: %d%%", p.Volume())

	spaceCount := innerW - 2 - len(timeStr) - len(volStr)
	if spaceCount < 1 {
		spaceCount = 1
	}
	line3 := " " + timeStr + strings.Repeat(" ", spaceCount) + volStr + " "

	innerLines := []string{line1, line2, line3}
	return RenderBoxWithTitle(statusTag, innerLines, width, 5, styles)
}
