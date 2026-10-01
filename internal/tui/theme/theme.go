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
	TextColor       lipgloss.Color
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
		BorderColor:     lipgloss.Color("#4a6984"), // Slate Blue
		ProgressColor:   lipgloss.Color("#7aa2f7"), // Muted Indigo
		VisualizerColor: lipgloss.Color("#6272a4"), // Dusk Slate
		AlertColor:      lipgloss.Color("#e06c75"), // Muted Terracotta
		TextColor:       lipgloss.Color("#abb2bf"), // Soft Silver Gray
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#1f2430"), // Deep Charcoal Slate
		SelectedFgColor: lipgloss.Color("#89ddff"), // Pale Ice
		VizBaseColor:    lipgloss.Color("#3b526d"), // Deep Slate
		VizMidColor:     lipgloss.Color("#55708f"), // Muted Steel
		VizHighColor:    lipgloss.Color("#7aa2f7"), // Soft Indigo
		VizPeakColor:    lipgloss.Color("#cbd5e1"), // Muted Silver
	},
	{
		Name:            "Nord",
		BorderColor:     lipgloss.Color("#81a1c1"), // Nord Frost Blue
		ProgressColor:   lipgloss.Color("#88c0d0"), // Nord Frost Cyan
		VisualizerColor: lipgloss.Color("#81a1c1"), // Nord Frost
		AlertColor:      lipgloss.Color("#bf616a"), // Nord Aurora Red
		TextColor:       lipgloss.Color("#d8dee9"), // Nord Snow Storm
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#2e3440"), // Nord Polar Night
		SelectedFgColor: lipgloss.Color("#eceff4"), // Nord Pure Snow
		VizBaseColor:    lipgloss.Color("#4c566a"), // Polar Slate
		VizMidColor:     lipgloss.Color("#5e81ac"), // Deep Frost
		VizHighColor:    lipgloss.Color("#81a1c1"), // Frost Blue
		VizPeakColor:    lipgloss.Color("#eceff4"), // Snow Frost
	},
	{
		Name:            "Matrix",
		BorderColor:     lipgloss.Color("#3f6652"), // Muted Pine
		ProgressColor:   lipgloss.Color("#589e74"), // Vintage Phosphor Green
		VisualizerColor: lipgloss.Color("#488b64"), // Phosphor Sage
		AlertColor:      lipgloss.Color("#d19a66"), // Warm Muted Amber
		TextColor:       lipgloss.Color("#87af87"), // Soft Sage
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#182a1f"), // Deep Forest
		SelectedFgColor: lipgloss.Color("#85d49b"), // Pale Mint
		VizBaseColor:    lipgloss.Color("#2b4e39"), // Deep Moss
		VizMidColor:     lipgloss.Color("#488b64"), // Sage Green
		VizHighColor:    lipgloss.Color("#6eb38a"), // Light Sage
		VizPeakColor:    lipgloss.Color("#b8e2c8"), // Pale Mint Glow
	},
	{
		Name:            "HyDE",
		BorderColor:     lipgloss.Color("#7d6b99"), // Muted Dusk Violet
		ProgressColor:   lipgloss.Color("#b48ead"), // Muted Heather Rose
		VisualizerColor: lipgloss.Color("#9a7ecc"), // Soft Lavender
		AlertColor:      lipgloss.Color("#e0a38d"), // Soft Peach
		TextColor:       lipgloss.Color("#b3a7c6"), // Muted Wisteria
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#282333"), // Deep Twilight
		SelectedFgColor: lipgloss.Color("#e5d9f2"), // Light Lavender
		VizBaseColor:    lipgloss.Color("#504068"), // Deep Plum
		VizMidColor:     lipgloss.Color("#7d6b99"), // Muted Violet
		VizHighColor:    lipgloss.Color("#b48ead"), // Heather Rose
		VizPeakColor:    lipgloss.Color("#ede7f6"), // Soft Milk Lavender
	},
	{
		Name:            "Gruvbox",
		BorderColor:     lipgloss.Color("#7c6f64"), // Muted Warm Gray
		ProgressColor:   lipgloss.Color("#d79921"), // Warm Desert Gold
		VisualizerColor: lipgloss.Color("#b16286"), // Muted Plum
		AlertColor:      lipgloss.Color("#cc241d"), // Muted Rust
		TextColor:       lipgloss.Color("#d5c4a1"), // Warm Sand
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#3c3836"), // Dark Espresso
		SelectedFgColor: lipgloss.Color("#fabd2f"), // Soft Gold
		VizBaseColor:    lipgloss.Color("#504945"), // Deep Umber
		VizMidColor:     lipgloss.Color("#b16286"), // Muted Wine
		VizHighColor:    lipgloss.Color("#d79921"), // Desert Gold
		VizPeakColor:    lipgloss.Color("#ebdbb2"), // Parchment
	},
	{
		Name:            "Slate",
		BorderColor:     lipgloss.Color("#64748b"), // Slate 500
		ProgressColor:   lipgloss.Color("#94a3b8"), // Slate 400
		VisualizerColor: lipgloss.Color("#cbd5e1"), // Slate 300
		AlertColor:      lipgloss.Color("#f87171"), // Muted Coral
		TextColor:       lipgloss.Color("#94a3b8"), // Muted Slate
		BgColor:         lipgloss.Color(""),
		SelectedBgColor: lipgloss.Color("#1e293b"), // Dark Slate
		SelectedFgColor: lipgloss.Color("#f8fafc"), // Off White
		VizBaseColor:    lipgloss.Color("#334155"), // Slate 700
		VizMidColor:     lipgloss.Color("#64748b"), // Slate 500
		VizHighColor:    lipgloss.Color("#94a3b8"), // Slate 400
		VizPeakColor:    lipgloss.Color("#e2e8f0"), // Soft Ash White
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
