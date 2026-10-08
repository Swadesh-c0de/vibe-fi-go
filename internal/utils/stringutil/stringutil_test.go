package stringutil

import (
	"testing"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		sec      float64
		expected string
	}{
		{0, "00:00"},
		{5, "00:05"},
		{65, "01:05"},
		{3665, "01:01:05"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.sec)
		if got != tt.expected {
			t.Errorf("FormatDuration(%.1f) = %q, expected %q", tt.sec, got, tt.expected)
		}
	}
}

func TestParseDuration(t *testing.T) {
	if sec := ParseDuration("03:45"); sec != 225.0 {
		t.Errorf("expected 225, got %.1f", sec)
	}
	if sec := ParseDuration("01:10:05"); sec != 4205.0 {
		t.Errorf("expected 4205, got %.1f", sec)
	}
}

func TestCleanTrackTitle(t *testing.T) {
	artist, track := CleanTrackTitle("Daft Punk - Get Lucky (Official Video)")
	if artist != "Daft Punk" || track != "Get Lucky" {
		t.Errorf("expected 'Daft Punk' and 'Get Lucky', got %q and %q", artist, track)
	}

	_, track2 := CleanTrackTitle("Bohemian Rhapsody [Lyrics]")
	if track2 != "Bohemian Rhapsody" {
		t.Errorf("expected 'Bohemian Rhapsody', got %q", track2)
	}
}

func TestFuzzyMatch(t *testing.T) {
	if !FuzzyMatch("chill", "Chillhop Daydreams") {
		t.Errorf("expected true")
	}
	if !FuzzyMatch("chd", "Chillhop Daydreams") {
		t.Errorf("expected true")
	}
	if FuzzyMatch("xyz", "Chillhop Daydreams") {
		t.Errorf("expected false")
	}
}

func TestDevanagariWidth(t *testing.T) {
	tests := []struct {
		str      string
		expected int
	}{
		{"छोड़ के न चल पड़य तू", 15},
		{"तन्ने मेरी याद, आवैगी", 14},
		{"कण कड़े रात खवेगी", 12},
		{"य ते मन्ने मर खंव रै", 15},
		{"स्यूं मन्ने, छोड़ के गय?", 16},
		{"Hello World", 11},
	}

	for _, tt := range tests {
		got := Width(tt.str)
		if got != tt.expected {
			t.Errorf("Width(%q) = %d, expected %d", tt.str, got, tt.expected)
		}
	}

	// Test PadRight
	padded := PadRight("छोड़ के न चल पड़य तू", 20)
	if Width(padded) != 20 {
		t.Errorf("PadRight width = %d, expected 20", Width(padded))
	}

	// Test PadCenter
	centered := PadCenter("छोड़ के न चल पड़य तू", 25)
	if Width(centered) != 25 {
		t.Errorf("PadCenter width = %d, expected 25", Width(centered))
	}
}

