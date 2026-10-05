package views

import (
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/tui/theme"
)

// RenderFullscreenLyricsView renders the lyrics viewer spanning the entire main area.
func RenderFullscreenLyricsView(width, height int, p player.AudioPlayer, lyricsData lyrics.LyricsData, lyricsLoading bool, animFrame int, lyricsScrollOffset int, autoScroll bool, styles theme.Styles) (string, int, bool) {
	lyricsLines, newOffset, newAutoScroll := renderLyricsBody(width-2, height-2, p, lyricsData, lyricsLoading, animFrame, lyricsScrollOffset, autoScroll, styles)
	return components.RenderBoxWithTitle("LYRICS", lyricsLines, width, height, styles), newOffset, newAutoScroll
}
