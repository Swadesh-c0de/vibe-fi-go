package library

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"vibe-fi/internal/utils/stringutil"
)

// LibraryItem represents a file or folder in the library.
type LibraryItem struct {
	Name        string
	Path        string
	Duration    string
	IsDirectory bool
}

var supportedExtensions = map[string]bool{
	".flac": true,
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
	".ogg":  true,
	".opus": true,
	".aac":  true,
	".alac": true,
	".aiff": true,
	".webm": true,
}

// Library handles file crawler and music directory browsing.
type Library struct {
	mu            sync.RWMutex
	durationCache map[string]string
}

// NewLibrary creates a new Library instance.
func NewLibrary() *Library {
	return &Library{
		durationCache: make(map[string]string),
	}
}

// IsAudioFile returns true if the file extension is a supported audio format.
func IsAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedExtensions[ext]
}

// GetHomeMusicDir resolves standard user music directory.
func (l *Library) GetHomeMusicDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}

	musicDir := filepath.Join(home, "Music")
	if fi, err := os.Stat(musicDir); err == nil && fi.IsDir() {
		return musicDir
	}

	return home
}

// ListDirectory lists folders and audio files in a given directory.
func (l *Library) ListDirectory(dirPath string) ([]LibraryItem, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var dirs []LibraryItem
	var files []LibraryItem

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue // skip hidden files
		}
		fullPath := filepath.Join(dirPath, name)

		if entry.IsDir() {
			dirs = append(dirs, LibraryItem{
				Name:        name,
				Path:        fullPath,
				IsDirectory: true,
			})
		} else if IsAudioFile(name) {
			l.mu.RLock()
			dur := l.durationCache[fullPath]
			l.mu.RUnlock()

			files = append(files, LibraryItem{
				Name:        name,
				Path:        fullPath,
				Duration:    dur,
				IsDirectory: false,
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return append(dirs, files...), nil
}

// SetCachedDuration records a duration for an audio file.
func (l *Library) SetCachedDuration(path, duration string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.durationCache[path] = duration
}

// Search filters items in a directory matching a query pattern.
func (l *Library) Search(dirPath, query string) ([]LibraryItem, error) {
	items, err := l.ListDirectory(dirPath)
	if err != nil {
		return nil, err
	}
	if query == "" {
		return items, nil
	}
	var filtered []LibraryItem
	for _, it := range items {
		if stringutil.FuzzyMatch(query, it.Name) {
			filtered = append(filtered, it)
		}
	}
	return filtered, nil
}
