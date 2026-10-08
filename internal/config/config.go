package config

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	dirsMu      sync.RWMutex
	ensuredDirs = make(map[string]bool)
)

func ensureDir(dir string) string {
	dirsMu.RLock()
	if ensuredDirs[dir] {
		dirsMu.RUnlock()
		return dir
	}
	dirsMu.RUnlock()

	dirsMu.Lock()
	defer dirsMu.Unlock()
	if !ensuredDirs[dir] {
		_ = os.MkdirAll(dir, 0755)
		ensuredDirs[dir] = true
	}
	return dir
}

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
	return ensureDir(dir)
}

// GetPlaylistsDir returns ~/.vibe-fi/playlists.
func GetPlaylistsDir() string {
	dir := filepath.Join(GetVibeDir(), "playlists")
	return ensureDir(dir)
}

// GetCacheDir returns ~/.vibe-fi/cache.
func GetCacheDir() string {
	dir := filepath.Join(GetVibeDir(), "cache")
	return ensureDir(dir)
}

// GetLyricsCacheDir returns ~/.vibe-fi/cache/lyrics.
func GetLyricsCacheDir() string {
	dir := filepath.Join(GetCacheDir(), "lyrics")
	return ensureDir(dir)
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
