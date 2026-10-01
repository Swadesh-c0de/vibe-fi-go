package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/tui/visualizer"
	"vibe-fi/internal/utils/stringutil"
)

// RenderPlaybackView renders the dual Visualizer (40%) and Lyrics (60%) layout.
func RenderPlaybackView(width, height int, p player.AudioPlayer, viz *visualizer.Visualizer, lyricsData lyrics.LyricsData, lyricsScrollOffset int, autoScroll bool, styles theme.Styles) (string, int) {
	if height < 6 {
		height = 6
	}

	vizH := int(float64(height) * 0.40)
	if vizH < 3 {
		vizH = 3
	}
	lyricsH := height - vizH
	if lyricsH < 3 {
		lyricsH = 3
	}

	// Top: Visualizer Box
	vizHeader := viz.RenderHeader(p)
	vizBody := viz.RenderBody(width-2, vizH-2, p, styles)
	vizBox := components.RenderBoxWithTitle(vizHeader, vizBody, width, vizH, styles)

	// Bottom: Lyrics Box
	lyricsLines, newOffset := renderLyricsBody(width-2, lyricsH-2, p, lyricsData, lyricsScrollOffset, autoScroll, styles)
	lyricsBox := components.RenderBoxWithTitle("LYRICS", lyricsLines, width, lyricsH, styles)

	return vizBox + "\n" + lyricsBox, newOffset
}

func renderLyricsBody(textW, textH int, p player.AudioPlayer, data lyrics.LyricsData, offset int, autoScroll bool, styles theme.Styles) ([]string, int) {
	if textW <= 0 || textH <= 0 {
		return nil, offset
	}

	out := make([]string, textH)
	for i := range out {
		out[i] = strings.Repeat(" ", textW)
	}

	if data.HasSynced && len(data.SyncedLyrics) > 0 {
		pos := p.Position()
		activeIdx := -1
		for i, line := range data.SyncedLyrics {
			if line.Timestamp <= pos {
				activeIdx = i
			} else {
				break
			}
		}

		if autoScroll && activeIdx != -1 {
			targetOffset := activeIdx - (textH / 2)
			if targetOffset < 0 {
				targetOffset = 0
			}
			offset = targetOffset
		}

		for i := 0; i < textH; i++ {
			idx := i + offset
			if idx >= len(data.SyncedLyrics) {
				break
			}

			lineText := data.SyncedLyrics[idx].Text
			isActive := (idx == activeIdx)
			if isActive {
				lineText = "> " + lineText
			}

			if lipgloss.Width(lineText) > textW-2 {
				lineText = ansi.Truncate(lineText, textW-5, "...")
			}
			displayLen := lipgloss.Width(lineText)

			leftPad := (textW - displayLen) / 2
			padRight := textW - leftPad - displayLen

			styledText := lineText
			if isActive {
				styledText = styles.ActiveSong.Render(lineText)
			}

			out[i] = stringutil.SafeRepeat(" ", leftPad) + styledText + stringutil.SafeRepeat(" ", padRight)
		}
	} else {
		// Plain lyrics or error
		if data.PlainLyrics == "" || strings.Contains(strings.ToLower(data.PlainLyrics), "not found") ||
			strings.Contains(strings.ToLower(data.PlainLyrics), "error") ||
			strings.Contains(strings.ToLower(data.PlainLyrics), "offline") {

			errMsg := data.PlainLyrics
			if errMsg == "" {
				errMsg = "Lyrics not found"
			}
			hint := "(Press 'S' to search YouTube)"

			midY := textH / 2
			errLen := lipgloss.Width(errMsg)
			errPad := (textW - errLen) / 2
			rErrPad := textW - errPad - errLen

			hintLen := lipgloss.Width(hint)
			hintPad := (textW - hintLen) / 2
			rHintPad := textW - hintPad - hintLen

			if midY < textH {
				out[midY] = stringutil.SafeRepeat(" ", errPad) + styles.StatusTitle.Render(errMsg) + stringutil.SafeRepeat(" ", rErrPad)
			}
			if midY+2 < textH {
				out[midY+2] = stringutil.SafeRepeat(" ", hintPad) + styles.StatusDim.Render(hint) + stringutil.SafeRepeat(" ", rHintPad)
			}
		} else {
			lines := strings.Split(data.PlainLyrics, "\n")
			for i := 0; i < textH; i++ {
				idx := i + offset
				if idx >= len(lines) {
					break
				}
				line := strings.TrimRight(lines[idx], "\r")
				if lipgloss.Width(line) > textW-2 {
					line = ansi.Truncate(line, textW-5, "...")
				}
				dispLen := lipgloss.Width(line)
				pad := (textW - dispLen) / 2
				rPad := textW - pad - dispLen
				out[i] = stringutil.SafeRepeat(" ", pad) + line + stringutil.SafeRepeat(" ", rPad)
			}
		}
	}

	return out, offset
}
