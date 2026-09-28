package components

import (
	"strings"
	"testing"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
)

func TestRenderBoxWithTitle(t *testing.T) {
	th := theme.GetTheme("Midnight")
	styles := theme.MakeStyles(th)

	widths := []int{20, 40, 80, 120}
	heights := []int{3, 5, 10, 20}

	for _, w := range widths {
		for _, h := range heights {
			inner := make([]string, h-2)
			for i := range inner {
				inner[i] = "Test content line"
			}
			box := RenderBoxWithTitle("TEST TITLE", inner, w, h, styles)
			lines := strings.Split(box, "\n")
			if len(lines) != h {
				t.Errorf("w=%d h=%d: expected %d lines, got %d", w, h, h, len(lines))
			}
		}
	}
}

func TestRenderStatusBar(t *testing.T) {
	mock := player.NewMockPlayer()
	th := theme.GetTheme("Midnight")
	styles := theme.MakeStyles(th)

	widths := []int{20, 30, 60, 100}
	for _, w := range widths {
		bar := RenderStatusBar(w, mock, styles)
		lines := strings.Split(bar, "\n")
		if len(lines) != 5 {
			t.Errorf("width %d: expected 5 lines, got %d", w, len(lines))
		}
	}
}

func TestRenderHelpBar(t *testing.T) {
	th := theme.GetTheme("Matrix")
	styles := theme.MakeStyles(th)

	widths := []int{20, 30, 60, 100}
	for _, w := range widths {
		bar := RenderHelpBar(w, ViewModePlayback, "", true, styles)
		lines := strings.Split(bar, "\n")
		if len(lines) != 3 {
			t.Errorf("width %d: expected 3 lines, got %d", w, len(lines))
		}

		// With message
		barMsg := RenderHelpBar(w, ViewModePlayback, "Important alert message that is very long", false, styles)
		linesMsg := strings.Split(barMsg, "\n")
		if len(linesMsg) != 3 {
			t.Errorf("width %d with msg: expected 3 lines, got %d", w, len(linesMsg))
		}
	}
}
