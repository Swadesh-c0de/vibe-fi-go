package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinThemes(t *testing.T) {
	if len(BuiltinThemes) != 6 {
		t.Fatalf("expected 6 builtin themes, got %d", len(BuiltinThemes))
	}
	th := GetTheme("Midnight")
	if th.Name != "Midnight" {
		t.Errorf("expected Midnight, got %s", th.Name)
	}
	// Case-insensitive lookup
	thLower := GetTheme("midnight")
	if thLower.Name != "Midnight" {
		t.Errorf("expected Midnight for lowercase lookup, got %s", thLower.Name)
	}
}

func TestParseThemeJSON(t *testing.T) {
	tj := ThemeJSON{
		Name:          "Custom Test",
		BorderColor:   "#123456",
		ProgressColor: "#654321",
	}
	th, ok := ParseThemeJSON(tj)
	if !ok {
		t.Fatal("expected successful parse")
	}
	if th.Name != "Custom Test" {
		t.Errorf("expected 'Custom Test', got '%s'", th.Name)
	}
	if string(th.BorderColor) != "#123456" {
		t.Errorf("expected BorderColor '#123456', got '%s'", th.BorderColor)
	}

	// Empty name should fail
	_, okEmpty := ParseThemeJSON(ThemeJSON{Name: "   "})
	if okEmpty {
		t.Error("expected empty name to fail")
	}
}

func TestLoadExternalThemesAndInit(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially themes.json does not exist. InitThemes should create sample themes.json and load it.
	InitThemes(tmpDir)

	targetPath := filepath.Join(tmpDir, "themes.json")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatalf("InitThemes should have created %s", targetPath)
	}

	// Themes should now contain builtin (6) + sample external (4) = 10
	if len(Themes) < 10 {
		t.Errorf("expected at least 10 themes after loading samples, got %d", len(Themes))
	}

	catppuccin := GetTheme("Catppuccin Mocha")
	if catppuccin.Name != "Catppuccin Mocha" {
		t.Errorf("expected Catppuccin Mocha to be loaded, got %s", catppuccin.Name)
	}

	// Test cycling themes
	next := CycleTheme("Catppuccin Mocha")
	if next.Name == "" {
		t.Error("CycleTheme should return a valid next theme")
	}
}
