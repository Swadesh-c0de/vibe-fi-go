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

	lines := viz.RenderBody(80, 10, mock, styles)
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}
	for i, l := range lines {
		if len(l) == 0 {
			t.Errorf("line %d is empty", i)
		}
	}

	header := viz.RenderHeader(mock)
	if len(header) == 0 {
		t.Errorf("header is empty")
	}
}
