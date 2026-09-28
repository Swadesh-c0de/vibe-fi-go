package stringutil

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// FormatDuration formats seconds into "MM:SS" or "HH:MM:SS".
func FormatDuration(seconds float64) string {
	if math.IsNaN(seconds) || seconds <= 0 {
		return "00:00"
	}
	totalSec := int(math.Floor(seconds))
	hrs := totalSec / 3600
	mins := (totalSec % 3600) / 60
	secs := totalSec % 60

	if hrs > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hrs, mins, secs)
	}
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

// ParseDuration converts "MM:SS" or "HH:MM:SS" string into seconds.
func ParseDuration(durStr string) float64 {
	durStr = strings.TrimSpace(durStr)
	if durStr == "" || durStr == "--:--" {
		return 0
	}
	parts := strings.Split(durStr, ":")
	if len(parts) == 2 {
		mins, _ := strconv.Atoi(parts[0])
		secs, _ := strconv.Atoi(parts[1])
		return float64(mins*60 + secs)
	} else if len(parts) == 3 {
		hrs, _ := strconv.Atoi(parts[0])
		mins, _ := strconv.Atoi(parts[1])
		secs, _ := strconv.Atoi(parts[2])
		return float64(hrs*3600 + mins*60 + secs)
	}
	return 0
}

// SanitizeText removes non-printable characters and trims extra spaces.
func SanitizeText(text string) string {
	var sb strings.Builder
	for _, r := range text {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			sb.WriteRune(r)
		}
	}
	return strings.TrimSpace(sb.String())
}

var noiseRegex = regexp.MustCompile(`(?i)\s*(\(|\[)(official\s+video|official\s+audio|lyrics?|visualizer|audio|hd|4k|mv|music\s+video|remastered|lyric\s+video)(\)|\])`)

// CleanTrackTitle parses artist and track from title string and removes YouTube noise.
func CleanTrackTitle(title string) (artist, track string) {
	cleaned := noiseRegex.ReplaceAllString(title, "")
	cleaned = strings.TrimSpace(cleaned)

	if idx := strings.Index(cleaned, " - "); idx != -1 {
		artist = strings.TrimSpace(cleaned[:idx])
		track = strings.TrimSpace(cleaned[idx+3:])
		return artist, track
	}

	return "", cleaned
}

// FuzzyMatch does a case-insensitive subsequence match.
func FuzzyMatch(pattern, text string) bool {
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)
	if pattern == "" {
		return true
	}
	pIdx := 0
	pRunes := []rune(pattern)
	for _, r := range text {
		if r == pRunes[pIdx] {
			pIdx++
			if pIdx == len(pRunes) {
				return true
			}
		}
	}
	return false
}

// IsURL checks whether the input string is an HTTP/HTTPS URL.
func IsURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}
