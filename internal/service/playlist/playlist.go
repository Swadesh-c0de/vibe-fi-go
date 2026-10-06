package playlist

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"vibe-fi/internal/config"
	"vibe-fi/internal/utils/stringutil"
)

// PlaylistSong represents a song record in a playlist.
type PlaylistSong struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Duration string `json:"duration"`
	Artist   string `json:"artist,omitempty"`
}

// Playlist represents an overview record of a playlist file.
type Playlist struct {
	Name      string
	Path      string
	SongCount int
}

// PlaylistManager provides thread-safe CRUD for text playlists.
type PlaylistManager struct {
	mu           sync.RWMutex
	playlistsDir string
}

// NewPlaylistManager creates and initializes PlaylistManager.
func NewPlaylistManager() *PlaylistManager {
	dir := config.GetPlaylistsDir()
	return &PlaylistManager{playlistsDir: dir}
}

func (m *PlaylistManager) playlistPath(name string) string {
	cleanName := strings.ReplaceAll(name, "/", "_")
	return filepath.Join(m.playlistsDir, cleanName+".txt")
}

func (m *PlaylistManager) jsonPlaylistPath(name string) string {
	cleanName := strings.ReplaceAll(name, "/", "_")
	return filepath.Join(m.playlistsDir, cleanName+".json")
}

// ParsePlaylistLine parses a single line from a text playlist file.
// Format is Title|URL|Duration. Since song titles frequently contain pipe characters
// (e.g. "Song | Trap | NCS" or "Artist - Title | Official Video"), parsing is done
// from right-to-left to ensure the URL and Duration are correctly extracted.
func ParsePlaylistLine(line string) (PlaylistSong, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return PlaylistSong{}, false
	}

	lastPipe := strings.LastIndex(line, "|")
	if lastPipe == -1 {
		return PlaylistSong{}, false
	}

	secondLast := strings.LastIndex(line[:lastPipe], "|")
	if secondLast != -1 {
		title := stringutil.SanitizeText(line[:secondLast])
		url := strings.TrimSpace(line[secondLast+1 : lastPipe])
		duration := strings.TrimSpace(line[lastPipe+1:])
		if duration == "" {
			duration = "--:--"
		}
		if title == "" && url == "" {
			return PlaylistSong{}, false
		}
		artist, _ := stringutil.CleanTrackTitle(title)
		return PlaylistSong{
			Title:    title,
			URL:      url,
			Duration: duration,
			Artist:   artist,
		}, true
	}

	// Single pipe fallback: Title|URL
	title := stringutil.SanitizeText(line[:lastPipe])
	url := strings.TrimSpace(line[lastPipe+1:])
	if title == "" && url == "" {
		return PlaylistSong{}, false
	}
	artist, _ := stringutil.CleanTrackTitle(title)
	return PlaylistSong{
		Title:    title,
		URL:      url,
		Duration: "--:--",
		Artist:   artist,
	}, true
}

// ListPlaylists lists all available playlists with their song counts.
func (m *PlaylistManager) ListPlaylists() []Playlist {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Playlist
	entries, err := os.ReadDir(m.playlistsDir)
	if err != nil {
		return result
	}

	seen := make(map[string]bool)

	// First scan .txt playlists (vibe-fi standard)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".txt") {
			pName := strings.TrimSuffix(name, ".txt")
			pPath := filepath.Join(m.playlistsDir, name)
			count := m.countSongs(pPath)
			seen[pName] = true
			result = append(result, Playlist{
				Name:      pName,
				Path:      pPath,
				SongCount: count,
			})
		}
	}

	// Fallback scan: include .json playlists if no corresponding .txt exists
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			pName := strings.TrimSuffix(name, ".json")
			if !seen[pName] {
				pPath := filepath.Join(m.playlistsDir, name)
				count := m.countSongs(pPath)
				seen[pName] = true
				result = append(result, Playlist{
					Name:      pName,
					Path:      pPath,
					SongCount: count,
				})
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result
}

func (m *PlaylistManager) countSongs(path string) int {
	if strings.HasSuffix(path, ".json") {
		data, err := os.ReadFile(path)
		if err != nil {
			return 0
		}
		var pj struct {
			Songs []struct {
				Title string `json:"title"`
			} `json:"songs"`
		}
		if err := json.Unmarshal(data, &pj); err == nil {
			return len(pj.Songs)
		}
		return 0
	}

	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if _, ok := ParsePlaylistLine(scanner.Text()); ok {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return count
	}
	return count
}

// GetPlaylistSongs loads all songs from a playlist.
func (m *PlaylistManager) GetPlaylistSongs(name string) []PlaylistSong {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getPlaylistSongsLocked(name)
}

func (m *PlaylistManager) getPlaylistSongsLocked(name string) []PlaylistSong {
	var songs []PlaylistSong
	pPath := m.playlistPath(name)
	f, err := os.Open(pPath)
	if err != nil {
		// Fallback to .json if .txt does not exist
		jsonPath := m.jsonPlaylistPath(name)
		data, jerr := os.ReadFile(jsonPath)
		if jerr == nil {
			var pj struct {
				Songs []struct {
					Title    string `json:"title"`
					URL      string `json:"url"`
					Duration string `json:"duration"`
				} `json:"songs"`
			}
			if err := json.Unmarshal(data, &pj); err == nil {
				for _, s := range pj.Songs {
					dur := s.Duration
					if dur == "" {
						dur = "--:--"
					}
					songs = append(songs, PlaylistSong{
						Title:    stringutil.SanitizeText(s.Title),
						URL:      s.URL,
						Duration: dur,
					})
				}
			}
		}
		return songs
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if song, ok := ParsePlaylistLine(scanner.Text()); ok {
			songs = append(songs, song)
		}
	}
	if err := scanner.Err(); err != nil {
		return songs
	}
	return songs
}

// CreatePlaylist creates an empty playlist file.
func (m *PlaylistManager) CreatePlaylist(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("playlist name cannot be empty")
	}
	pPath := m.playlistPath(name)
	if _, err := os.Stat(pPath); err == nil {
		return fmt.Errorf("playlist already exists")
	}

	f, err := os.Create(pPath)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// DeletePlaylist deletes a playlist file (.txt and .json fallback).
func (m *PlaylistManager) DeletePlaylist(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	txtErr := os.Remove(m.playlistPath(name))
	jsonErr := os.Remove(m.jsonPlaylistPath(name))
	if txtErr != nil && jsonErr != nil {
		return txtErr
	}
	return nil
}

// RenamePlaylist renames a playlist file.
func (m *PlaylistManager) RenamePlaylist(oldName, newName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldTxt := m.playlistPath(oldName)
	newTxt := m.playlistPath(newName)
	txtErr := os.Rename(oldTxt, newTxt)

	oldJSON := m.jsonPlaylistPath(oldName)
	newJSON := m.jsonPlaylistPath(newName)
	_ = os.Rename(oldJSON, newJSON)

	return txtErr
}

// AddSongToPlaylist appends a song to a playlist, checking for duplicates.
func (m *PlaylistManager) AddSongToPlaylist(playlistName string, song PlaylistSong) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.getPlaylistSongsLocked(playlistName)
	for _, s := range current {
		if s.URL == song.URL {
			return nil // Avoid duplicate URLs
		}
	}

	pPath := m.playlistPath(playlistName)
	f, err := os.OpenFile(pPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	dur := song.Duration
	if dur == "" {
		dur = "--:--"
	}
	line := fmt.Sprintf("%s|%s|%s\n", song.Title, song.URL, dur)
	_, err = f.WriteString(line)
	return err
}

// RemoveSongFromPlaylist deletes a song at a given index.
func (m *PlaylistManager) RemoveSongFromPlaylist(playlistName string, index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	songs := m.getPlaylistSongsLocked(playlistName)
	if index < 0 || index >= len(songs) {
		return fmt.Errorf("index out of range")
	}

	songs = append(songs[:index], songs[index+1:]...)

	pPath := m.playlistPath(playlistName)
	f, err := os.Create(pPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, s := range songs {
		dur := s.Duration
		if dur == "" {
			dur = "--:--"
		}
		w.WriteString(fmt.Sprintf("%s|%s|%s\n", s.Title, s.URL, dur))
	}
	return w.Flush()
}

// MoveSong moves a song from src playlist to dest playlist.
func (m *PlaylistManager) MoveSong(srcPlaylist string, srcIndex int, destPlaylist string) error {
	songs := m.GetPlaylistSongs(srcPlaylist)
	if srcIndex < 0 || srcIndex >= len(songs) {
		return fmt.Errorf("invalid song index")
	}
	song := songs[srcIndex]

	if err := m.AddSongToPlaylist(destPlaylist, song); err != nil {
		return err
	}
	return m.RemoveSongFromPlaylist(srcPlaylist, srcIndex)
}

// ExportToM3U exports a playlist to an M3U file.
func (m *PlaylistManager) ExportToM3U(playlistName string, outPath string) error {
	songs := m.GetPlaylistSongs(playlistName)
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("#EXTM3U\n")

	for _, s := range songs {
		sec := int(stringutil.ParseDuration(s.Duration))
		if sec <= 0 {
			sec = -1
		}
		w.WriteString(fmt.Sprintf("#EXTINF:%d,%s\n%s\n", sec, s.Title, s.URL))
	}
	return w.Flush()
}
