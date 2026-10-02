package views

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/tui/visualizer"
	"vibe-fi/internal/utils/stringutil"
)

// RenderPlaybackView renders the dual Visualizer (40%) and Lyrics (60%) layout.
func RenderPlaybackView(width, height int, p player.AudioPlayer, viz *visualizer.Visualizer, lyricsData lyrics.LyricsData, lyricsScrollOffset int, autoScroll bool, styles theme.Styles) (string, int, bool) {
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
	lyricsLines, newOffset, newAutoScroll := renderLyricsBody(width-2, lyricsH-2, p, lyricsData, lyricsScrollOffset, autoScroll, styles)
	lyricsBox := components.RenderBoxWithTitle("LYRICS", lyricsLines, width, lyricsH, styles)

	return vizBox + "\n" + lyricsBox, newOffset, newAutoScroll
}

func renderLyricsBody(textW, textH int, p player.AudioPlayer, data lyrics.LyricsData, offset int, autoScroll bool, styles theme.Styles) ([]string, int, bool) {
	if textW <= 0 || textH <= 0 {
		return nil, offset, autoScroll
	}

	out := make([]string, textH)
	newAutoScroll := autoScroll

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

		if activeIdx != -1 {
			// If auto-scroll is on, OR if the active playing line reached or passed the visible bounds:
			// Automatically scroll to center the active line and resume auto-scroll!
			if autoScroll || activeIdx >= offset+textH-1 || activeIdx < offset {
				targetOffset := activeIdx - (textH / 2)
				if targetOffset < 0 {
					targetOffset = 0
				}
				offset = targetOffset
				newAutoScroll = true
			}
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

			if stringutil.Width(lineText) > textW-2 {
				lineText = ansi.Truncate(lineText, textW-5, "...")
			}
			displayLen := stringutil.Width(lineText)

			leftPad := (textW - displayLen) / 2
			if leftPad < 0 {
				leftPad = 0
			}

			styledText := lineText
			if isActive {
				styledText = styles.ActiveSong.Render(lineText)
			}

			out[i] = stringutil.SafeRepeat(" ", leftPad) + styledText
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
			errLen := stringutil.Width(errMsg)
			errPad := (textW - errLen) / 2
			if errPad < 0 {
				errPad = 0
			}

			hintLen := stringutil.Width(hint)
			hintPad := (textW - hintLen) / 2
			if hintPad < 0 {
				hintPad = 0
			}

			if midY < textH {
				out[midY] = stringutil.SafeRepeat(" ", errPad) + styles.StatusTitle.Render(errMsg)
			}
			if midY+2 < textH {
				out[midY+2] = stringutil.SafeRepeat(" ", hintPad) + styles.StatusDim.Render(hint)
			}
		} else {
			lines := strings.Split(data.PlainLyrics, "\n")

			// Auto-scroll plain lyrics proportionally based on track position
			dur := p.Duration()
			if dur > 0 && len(lines) > textH {
				prog := p.Position() / dur
				if prog < 0 {
					prog = 0
				} else if prog > 1 {
					prog = 1
				}
				maxOffset := len(lines) - textH
				targetOffset := int(prog * float64(maxOffset))
				if autoScroll || targetOffset >= offset+textH-1 || targetOffset < offset {
					offset = targetOffset
					newAutoScroll = true
				}
			}

			for i := 0; i < textH; i++ {
				idx := i + offset
				if idx >= len(lines) {
					break
				}
				line := strings.TrimRight(lines[idx], "\r\n")
				if stringutil.Width(line) > textW-2 {
					line = ansi.Truncate(line, textW-5, "...")
				}
				dispLen := stringutil.Width(line)
				pad := (textW - dispLen) / 2
				if pad < 0 {
					pad = 0
				}
				out[i] = stringutil.SafeRepeat(" ", pad) + line
			}
		}
	}

	return out, offset, newAutoScroll
}
