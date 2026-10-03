package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
	"vibe-fi/internal/utils/stringutil"
)

// formatAudioBadge resolves source audio codec and bitrate/samplerate for the status bar.
func formatAudioBadge(p player.AudioPlayer) string {
	if p.IsIdle() {
		return "[IDLE]"
	}
	if p.IsLoading() {
		return "[FETCHING]"
	}
	if p.IsBuffering() {
		return "[BUFFERING]"
	}

	codec := strings.ToUpper(p.GetMetadata("audio-codec-name"))
	if codec == "" {
		codec = strings.ToUpper(p.GetMetadata("file-format"))
		if strings.Contains(codec, ",") {
			codec = strings.Split(codec, ",")[0]
		}
	}
	if codec == "MATROSKA,WEBM" {
		codec = "OPUS"
	}

	bitrateStr := p.GetMetadata("audio-bitrate")
	if bitrateStr != "" {
		if br, err := strconv.Atoi(bitrateStr); err == nil && br > 0 {
			kbps := br / 1000
			if codec != "" {
				return fmt.Sprintf("[%s %dk]", codec, kbps)
			}
			return fmt.Sprintf("[%dkbps]", kbps)
		}
	}

	rateStr := p.GetMetadata("audio-params/samplerate")
	if rateStr != "" {
		if rate, err := strconv.Atoi(rateStr); err == nil && rate > 0 {
			rateK := float64(rate) / 1000.0
			if codec != "" {
				return fmt.Sprintf("[%s %.1fk]", codec, rateK)
			}
			return fmt.Sprintf("[%.1fkHz]", rateK)
		}
	}

	if codec != "" {
		return fmt.Sprintf("[%s]", codec)
	}

	return "[AUDIO]"
}

// RenderStatusBar renders the 5-line modern status bar at the bottom.
func RenderStatusBar(width int, p player.AudioPlayer, styles theme.Styles) string {
	statusTag := "NOW PLAYING"
	if p.IsLoading() {
		statusTag = "NOW PLAYING [FETCHING]"
	} else if p.IsBuffering() {
		statusTag = "NOW PLAYING [BUFFERING]"
	} else if p.IsPaused() {
		statusTag = "NOW PLAYING [⏸]"
	} else if p.IsPlaying() {
		statusTag = "NOW PLAYING [▶]"
	} else if p.IsIdle() {
		statusTag = "NOW PLAYING [IDLE]"
	}

	innerW := width - 2
	if innerW < 10 {
		innerW = 10
	}

	// Line 1: Centered Track Title
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

	if stringutil.Width(title) > innerW-2 {
		title = ansi.Truncate(title, innerW-5, "...")
	}
	tWidth := stringutil.Width(title)
	titleLeft := (innerW - tWidth) / 2
	titleRight := innerW - titleLeft - tWidth
	line1 := stringutil.SafeRepeat(" ", titleLeft) + styles.StatusTitle.Render(title) + stringutil.SafeRepeat(" ", titleRight)

	// Line 2: Modern Unicode Progress Bar with Inline Timestamps
	pos := p.Position()
	dur := p.Duration()

	var line2 string
	if dur > 0 {
		posStr := stringutil.FormatDuration(pos)
		durStr := stringutil.FormatDuration(dur)
		headText := " " + posStr + " "
		tailText := " " + durStr + " "

		railW := innerW - len(headText) - len(tailText)
		if railW >= 4 {
			prog := pos / dur
			if prog < 0 {
				prog = 0
			} else if prog > 1 {
				prog = 1
			}

			activeRail := railW - 1
			filled := int(prog * float64(activeRail))
			if filled > activeRail {
				filled = activeRail
			}
			if filled < 0 {
				filled = 0
			}
			empty := activeRail - filled

			playedStr := styles.ProgressBar.Render(strings.Repeat("━", filled))
			beadStr := styles.StatusTitle.Render("●")
			remainStr := styles.StatusDim.Render(strings.Repeat("─", empty))

			line2 = " " + styles.StatusDim.Render(posStr) + " " + playedStr + beadStr + remainStr + " " + styles.StatusDim.Render(durStr) + " "
		} else {
			line2 = stringutil.SafeRepeat(" ", innerW)
		}
	} else if p.IsLoading() || p.IsBuffering() {
		hint := "Connecting to audio stream..."
		if p.IsBuffering() {
			hint = "Buffering audio stream..."
		}
		hWidth := stringutil.Width(hint)
		hPad := (innerW - hWidth) / 2
		if hPad < 0 {
			hPad = 0
		}
		rPad := innerW - hPad - hWidth
		if rPad < 0 {
			rPad = 0
		}
		line2 = stringutil.SafeRepeat(" ", hPad) + styles.ProgressBar.Render(hint) + stringutil.SafeRepeat(" ", rPad)
	} else {
		line2 = stringutil.SafeRepeat(" ", innerW)
	}

	// Line 3: Audio Codec Badge on left, Volume on right
	badge := formatAudioBadge(p)
	badgeRendered := styles.StatusDim.Render(badge)
	volStr := fmt.Sprintf("Vol: %d%%", p.Volume())
	volRendered := styles.StatusDim.Render(volStr)

	bWidth := stringutil.Width(badge)
	vWidth := stringutil.Width(volStr)
	spaceCount := innerW - 2 - bWidth - vWidth
	if spaceCount < 1 {
		spaceCount = 1
	}
	line3 := " " + badgeRendered + stringutil.SafeRepeat(" ", spaceCount) + volRendered + " "

	innerLines := []string{line1, line2, line3}
	return RenderBoxWithTitle(statusTag, innerLines, width, 5, styles)
}
