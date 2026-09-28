package visualizer

import (
	"testing"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
)

func TestVisualizerRender(t *testing.T) {
	mock := player.NewMockPlayer()
	_ = mock.Load("test.mp3", "replace")
	_ = mock.Play()

	th := theme.GetTheme("Midnight")
	styles := theme.MakeStyles(th)
	viz := NewVisualizer()

	modes := []VisualizerMode{ModeCavaWave, ModeNeonFlame, ModeStereoBars}
	for _, m := range modes {
		lines := viz.RenderBody(80, 10, mock, m, styles)
		if len(lines) != 10 {
			t.Errorf("mode %s: expected 10 lines, got %d", m.String(), len(lines))
		}
		for i, l := range lines {
			if len(l) == 0 {
				t.Errorf("mode %s line %d is empty", m.String(), i)
			}
		}

		header := viz.RenderHeader(mock, m)
		if len(header) == 0 {
			t.Errorf("mode %s header is empty", m.String())
		}
	}
}
