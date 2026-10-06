package config

import (
	"os"
	"path/filepath"
)

const (
	Version           = "2.0.0"
	DefaultTheme      = "Midnight"
	DefaultVisualizer = 0 // Cava Wave
)

// GetVibeDir returns ~/.vibe-fi, creating it if it doesn't exist.
func GetVibeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	dir := filepath.Join(home, ".vibe-fi")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// GetPlaylistsDir returns ~/.vibe-fi/playlists.
func GetPlaylistsDir() string {
	dir := filepath.Join(GetVibeDir(), "playlists")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// GetCacheDir returns ~/.vibe-fi/cache.
func GetCacheDir() string {
	dir := filepath.Join(GetVibeDir(), "cache")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// GetLyricsCacheDir returns ~/.vibe-fi/cache/lyrics.
func GetLyricsCacheDir() string {
	dir := filepath.Join(GetCacheDir(), "lyrics")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// GetBottleDir returns $VIBE_BOTTLE_DIR or ~/.vibe-fi/bottle.
func GetBottleDir() string {
	if envDir := os.Getenv("VIBE_BOTTLE_DIR"); envDir != "" {
		return envDir
	}
	dir := filepath.Join(GetVibeDir(), "bottle")
	return dir
}

// GetBottleBinDir returns the bin directory inside bottle.
func GetBottleBinDir() string {
	return filepath.Join(GetBottleDir(), "bin")
}
