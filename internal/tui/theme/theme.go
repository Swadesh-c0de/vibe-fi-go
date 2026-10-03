package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme encapsulates a complete color palette for the player.
type Theme struct {
	Name            string         `json:"name"`
	BorderColor     lipgloss.Color `json:"borderColor"`
	ProgressColor   lipgloss.Color `json:"progressColor"`
	AlertColor      lipgloss.Color `json:"alertColor"`
	TextColor       lipgloss.Color `json:"textColor"`
	SelectedBgColor lipgloss.Color `json:"selectedBgColor"`
	SelectedFgColor lipgloss.Color `json:"selectedFgColor"`

	// Multi-tier gradient colors for visualizer
	VizBaseColor lipgloss.Color `json:"vizBaseColor"`
	VizMidColor  lipgloss.Color `json:"vizMidColor"`
	VizHighColor lipgloss.Color `json:"vizHighColor"`
}

// ThemeJSON represents the external JSON file format for custom themes.
type ThemeJSON struct {
	Name            string `json:"name"`
	BorderColor     string `json:"borderColor"`
	ProgressColor   string `json:"progressColor"`
	AlertColor      string `json:"alertColor"`
	TextColor       string `json:"textColor"`
	SelectedBgColor string `json:"selectedBgColor"`
	SelectedFgColor string `json:"selectedFgColor"`
	VizBaseColor    string `json:"vizBaseColor"`
	VizMidColor     string `json:"vizMidColor"`
	VizHighColor    string `json:"vizHighColor"`
}

// BuiltinThemes defines the hardcoded default theme set.
var BuiltinThemes = []Theme{
	{
		Name:            "Midnight",
		BorderColor:     lipgloss.Color("#4e657e"), // Soft Dusty Slate Blue
		ProgressColor:   lipgloss.Color("#6c8cad"), // Muted Steel Blue
		AlertColor:      lipgloss.Color("#b86868"), // Soft Terracotta
		TextColor:       lipgloss.Color("#9eb0c2"), // Soft Morning Mist
		SelectedBgColor: lipgloss.Color("#1d2530"), // Deep Midnight Slate
		SelectedFgColor: lipgloss.Color("#c4d5e7"), // Soft Pale Cloud
		VizBaseColor:    lipgloss.Color("#2d3d4f"), // Deep Ambient Dusk
		VizMidColor:     lipgloss.Color("#4a637d"), // Soft Slate
		VizHighColor:    lipgloss.Color("#6c8cad"), // Gentle Steel Mist
	},
	{
		Name:            "Nord",
		BorderColor:     lipgloss.Color("#647b91"), // Nord Polar Slate
		ProgressColor:   lipgloss.Color("#7aa5b3"), // Muted Frost Aqua
		AlertColor:      lipgloss.Color("#b06b72"), // Soft Aurora Red
		TextColor:       lipgloss.Color("#b5c4d4"), // Soft Polar Fog
		SelectedBgColor: lipgloss.Color("#242b35"), // Deep Polar Night
		SelectedFgColor: lipgloss.Color("#d8e2ec"), // Soft Snow White
		VizBaseColor:    lipgloss.Color("#3b4856"), // Deep Arctic Base
		VizMidColor:     lipgloss.Color("#52687a"), // Muted Ice
		VizHighColor:    lipgloss.Color("#7aa5b3"), // Gentle Frost Aqua
	},
	{
		Name:            "Matrix",
		BorderColor:     lipgloss.Color("#456353"), // Soft Damp Moss
		ProgressColor:   lipgloss.Color("#6b947c"), // Muted Matcha Sage
		AlertColor:      lipgloss.Color("#b58362"), // Warm Muted Amber
		TextColor:       lipgloss.Color("#8fa89b"), // Pale Dried Sage
		SelectedBgColor: lipgloss.Color("#18241d"), // Deep Pine Dusk
		SelectedFgColor: lipgloss.Color("#b2cfbf"), // Soft Celadon Mist
		VizBaseColor:    lipgloss.Color("#2a3d33"), // Deep Forest Floor
		VizMidColor:     lipgloss.Color("#486b58"), // Soft Moss Green
		VizHighColor:    lipgloss.Color("#6b947c"), // Gentle Matcha Sage
	},
	{
		Name:            "HyDE",
		BorderColor:     lipgloss.Color("#695c78"), // Muted Dusk Plum
		ProgressColor:   lipgloss.Color("#9884ac"), // Soft Dusty Lilac
		AlertColor:      lipgloss.Color("#b87480"), // Muted Mauve Red
		TextColor:       lipgloss.Color("#ada0bd"), // Soft Heather Mist
		SelectedBgColor: lipgloss.Color("#221c29"), // Deep Twilight Velvet
		SelectedFgColor: lipgloss.Color("#d6cbe3"), // Pale Lavender Dew
		VizBaseColor:    lipgloss.Color("#3b3047"), // Deep Plum Shadow
		VizMidColor:     lipgloss.Color("#5f4f70"), // Soft Wisteria
		VizHighColor:    lipgloss.Color("#9884ac"), // Dusty Lilac
	},
	{
		Name:            "Gruvbox",
		BorderColor:     lipgloss.Color("#6c6258"), // Muted Driftwood
		ProgressColor:   lipgloss.Color("#b89660"), // Soft Golden Sand
		AlertColor:      lipgloss.Color("#b05a54"), // Muted Baked Clay
		TextColor:       lipgloss.Color("#b8ab97"), // Warm Oat Parchment
		SelectedBgColor: lipgloss.Color("#292522"), // Dark Espresso Warmth
		SelectedFgColor: lipgloss.Color("#dbcdb8"), // Pale Flaxen Cream
		VizBaseColor:    lipgloss.Color("#403934"), // Deep Earth
		VizMidColor:     lipgloss.Color("#736557"), // Warm Sandstone
		VizHighColor:    lipgloss.Color("#b89660"), // Soft Golden Sand
	},
	{
		Name:            "Slate",
		BorderColor:     lipgloss.Color("#5c6773"), // Muted Slate 600
		ProgressColor:   lipgloss.Color("#7d8b99"), // Soft Slate Gray
		AlertColor:      lipgloss.Color("#ab6d6d"), // Muted Rosewood
		TextColor:       lipgloss.Color("#9ba8b5"), // Soft Cloud Gray
		SelectedBgColor: lipgloss.Color("#1d2228"), // Dark Charcoal Slate
		SelectedFgColor: lipgloss.Color("#d0d9e2"), // Pale Fog Mist
		VizBaseColor:    lipgloss.Color("#303942"), // Deep Charcoal Shadow
		VizMidColor:     lipgloss.Color("#4f5c69"), // Mid Slate
		VizHighColor:    lipgloss.Color("#7d8b99"), // Soft Slate Gray
	},
}

// Themes holds the active set of themes (built-in + external custom).
var Themes = append([]Theme{}, BuiltinThemes...)

// SampleExternalThemes returns popular community themes for the default themes.json.
func SampleExternalThemes() []ThemeJSON {
	return []ThemeJSON{
		{
			Name:            "Catppuccin Mocha",
			BorderColor:     "#68768e",
			ProgressColor:   "#82a58b",
			AlertColor:      "#b8747a",
			TextColor:       "#aeb7cb",
			SelectedBgColor: "#222430",
			SelectedFgColor: "#d2daf0",
			VizBaseColor:    "#333748",
			VizMidColor:     "#596680",
			VizHighColor:    "#8ea3c4",
		},
		{
			Name:            "Tokyo Dusk",
			BorderColor:     "#556885",
			ProgressColor:   "#6897b0",
			AlertColor:      "#b56872",
			TextColor:       "#98a6bf",
			SelectedBgColor: "#1c2230",
			SelectedFgColor: "#c0cde6",
			VizBaseColor:    "#2b364a",
			VizMidColor:     "#485c7b",
			VizHighColor:    "#6897b0",
		},
		{
			Name:            "Sage",
			BorderColor:     "#5f6e62",
			ProgressColor:   "#829c88",
			AlertColor:      "#ba7d6c",
			TextColor:       "#a7b8ab",
			SelectedBgColor: "#1f2621",
			SelectedFgColor: "#cbe0cf",
			VizBaseColor:    "#344237",
			VizMidColor:     "#596e5e",
			VizHighColor:    "#829c88",
		},
		{
			Name:            "Rose Pine",
			BorderColor:     "#7a6675",
			ProgressColor:   "#b88a95",
			AlertColor:      "#bd6272",
			TextColor:       "#bfa8b3",
			SelectedBgColor: "#231c22",
			SelectedFgColor: "#decad4",
			VizBaseColor:    "#3d2f39",
			VizMidColor:     "#6e5464",
			VizHighColor:    "#b88a95",
		},
	}
}

// ParseThemeJSON converts a ThemeJSON into a Theme with lipgloss colors.
func ParseThemeJSON(t ThemeJSON) (Theme, bool) {
	if strings.TrimSpace(t.Name) == "" {
		return Theme{}, false
	}
	// Fallback to sensible defaults if colors omitted
	def := BuiltinThemes[0]
	toColor := func(hex string, fallback lipgloss.Color) lipgloss.Color {
		hex = strings.TrimSpace(hex)
		if hex == "" {
			return fallback
		}
		return lipgloss.Color(hex)
	}

	return Theme{
		Name:            strings.TrimSpace(t.Name),
		BorderColor:     toColor(t.BorderColor, def.BorderColor),
		ProgressColor:   toColor(t.ProgressColor, def.ProgressColor),
		AlertColor:      toColor(t.AlertColor, def.AlertColor),
		TextColor:       toColor(t.TextColor, def.TextColor),
		SelectedBgColor: toColor(t.SelectedBgColor, def.SelectedBgColor),
		SelectedFgColor: toColor(t.SelectedFgColor, def.SelectedFgColor),
		VizBaseColor:    toColor(t.VizBaseColor, def.VizBaseColor),
		VizMidColor:     toColor(t.VizMidColor, def.VizMidColor),
		VizHighColor:    toColor(t.VizHighColor, def.VizHighColor),
	}, true
}

// LoadExternalThemes reads and parses custom themes from a JSON file.
func LoadExternalThemes(filePath string) ([]Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var jsonThemes []ThemeJSON
	if err := json.Unmarshal(data, &jsonThemes); err != nil {
		return nil, err
	}

	var loaded []Theme
	for _, jt := range jsonThemes {
		if th, ok := ParseThemeJSON(jt); ok {
			loaded = append(loaded, th)
		}
	}
	return loaded, nil
}

// RegisterThemes merges external themes into the active Themes list.
// If a theme name matches an existing one, it replaces it; otherwise it appends.
func RegisterThemes(custom []Theme) {
	for _, c := range custom {
		replaced := false
		for i, existing := range Themes {
			if strings.EqualFold(existing.Name, c.Name) {
				Themes[i] = c
				replaced = true
				break
			}
		}
		if !replaced {
			Themes = append(Themes, c)
		}
	}
}

// InitThemes initializes themes by resetting to built-ins and loading from config directory.
// If themes.json does not exist in configDir, it creates a default template with popular themes.
func InitThemes(configDir string) {
	Themes = append([]Theme{}, BuiltinThemes...)

	if configDir == "" {
		return
	}

	targetPath := filepath.Join(configDir, "themes.json")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// Write default sample themes
		samples := SampleExternalThemes()
		if data, err := json.MarshalIndent(samples, "", "  "); err == nil {
			_ = os.WriteFile(targetPath, data, 0644)
		}
	}

	if loaded, err := LoadExternalThemes(targetPath); err == nil && len(loaded) > 0 {
		RegisterThemes(loaded)
	}

	// Also check ~/.config/vibe-fi/themes.json if distinct from configDir
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		altPath := filepath.Join(home, ".config", "vibe-fi", "themes.json")
		if altPath != targetPath {
			if loaded, err := LoadExternalThemes(altPath); err == nil && len(loaded) > 0 {
				RegisterThemes(loaded)
			}
		}
	}
}

// GetTheme returns the theme matching name, or Midnight by default.
func GetTheme(name string) Theme {
	for _, t := range Themes {
		if strings.EqualFold(t.Name, name) {
			return t
		}
	}
	if len(Themes) > 0 {
		return Themes[0]
	}
	return BuiltinThemes[0]
}

// CycleTheme returns the next theme after current.
func CycleTheme(current string) Theme {
	if len(Themes) == 0 {
		return BuiltinThemes[0]
	}
	for i, t := range Themes {
		if strings.EqualFold(t.Name, current) {
			nextIdx := (i + 1) % len(Themes)
			return Themes[nextIdx]
		}
	}
	return Themes[0]
}
