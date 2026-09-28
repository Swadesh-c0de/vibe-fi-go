package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme encapsulates a complete color palette for the player.
type Theme struct {
	Name            string
	BorderColor     lipgloss.Color
	ProgressColor   lipgloss.Color
	VisualizerColor lipgloss.Color
	AlertColor      lipgloss.Color
	BgColor         lipgloss.Color
	SelectedBgColor lipgloss.Color
	SelectedFgColor lipgloss.Color

	// Multi-tier gradient colors for visualizer
	VizBaseColor lipgloss.Color
	VizMidColor  lipgloss.Color
	VizHighColor lipgloss.Color
	VizPeakColor lipgloss.Color
}

var Themes = []Theme{
	{
		Name:            "Midnight",
		BorderColor:     lipgloss.Color("#3b82f6"), // Blue / Deep Indigo
		ProgressColor:   lipgloss.Color("#d946ef"), // Magenta
		VisualizerColor: lipgloss.Color("#06b6d4"), // Electric Cyan
		AlertColor:      lipgloss.Color("#ef4444"), // Red
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#000000"),
		SelectedFgColor: lipgloss.Color("#06b6d4"),
		VizBaseColor:    lipgloss.Color("#06b6d4"), // Cyan
		VizMidColor:     lipgloss.Color("#d946ef"), // Magenta
		VizHighColor:    lipgloss.Color("#ef4444"), // Red
		VizPeakColor:    lipgloss.Color("#ffffff"), // White
	},
	{
		Name:            "Matrix",
		BorderColor:     lipgloss.Color("#22c55e"), // Terminal Green
		ProgressColor:   lipgloss.Color("#22c55e"), // Bright Green
		VisualizerColor: lipgloss.Color("#22c55e"), // Phosphor Green
		AlertColor:      lipgloss.Color("#ef4444"), // Red
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#000000"),
		SelectedFgColor: lipgloss.Color("#22c55e"),
		VizBaseColor:    lipgloss.Color("#16a34a"), // Deep Green
		VizMidColor:     lipgloss.Color("#22c55e"), // Bright Green
		VizHighColor:    lipgloss.Color("#ef4444"), // Red
		VizPeakColor:    lipgloss.Color("#ffffff"), // White
	},
	{
		Name:            "Nord",
		BorderColor:     lipgloss.Color("#88c0d0"), // Arctic Cyan
		ProgressColor:   lipgloss.Color("#81a1c1"), // Frost Blue
		VisualizerColor: lipgloss.Color("#eceff4"), // Clean Frost White
		AlertColor:      lipgloss.Color("#bf616a"), // Red
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#000000"),
		SelectedFgColor: lipgloss.Color("#88c0d0"),
		VizBaseColor:    lipgloss.Color("#eceff4"), // Frost White
		VizMidColor:     lipgloss.Color("#81a1c1"), // Cool Blue
		VizHighColor:    lipgloss.Color("#bf616a"), // Aurora Red
		VizPeakColor:    lipgloss.Color("#ffffff"), // Pure White
	},
	{
		Name:            "HyDE",
		BorderColor:     lipgloss.Color("#c084fc"), // Velvet Magenta
		ProgressColor:   lipgloss.Color("#22d3ee"), // Vivid Cyan
		VisualizerColor: lipgloss.Color("#c084fc"), // Violet
		AlertColor:      lipgloss.Color("#ffffff"), // White
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#000000"),
		SelectedFgColor: lipgloss.Color("#c084fc"),
		VizBaseColor:    lipgloss.Color("#c084fc"), // Velvet Magenta
		VizMidColor:     lipgloss.Color("#22d3ee"), // Vivid Cyan
		VizHighColor:    lipgloss.Color("#ffffff"), // White
		VizPeakColor:    lipgloss.Color("#ffffff"), // Pure White
	},
}

// GetTheme returns the theme matching name, or Midnight by default.
func GetTheme(name string) Theme {
	for _, t := range Themes {
		if t.Name == name {
			return t
		}
	}
	return Themes[0]
}

// CycleTheme returns the next theme after current.
func CycleTheme(current string) Theme {
	for i, t := range Themes {
		if t.Name == current {
			nextIdx := (i + 1) % len(Themes)
			return Themes[nextIdx]
		}
	}
	return Themes[0]
}
