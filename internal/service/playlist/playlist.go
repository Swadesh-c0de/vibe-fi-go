package playlist

import (
	"bufio"
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
	Title    string
	URL      string
	Duration string
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

// ListPlaylists lists all available playlists with their song counts.
func (m *PlaylistManager) ListPlaylists() []Playlist {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Playlist
	entries, err := os.ReadDir(m.playlistsDir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		pName := strings.TrimSuffix(entry.Name(), ".txt")
		pPath := filepath.Join(m.playlistsDir, entry.Name())
		count := m.countSongs(pPath)
		result = append(result, Playlist{
			Name:      pName,
			Path:      pPath,
			SongCount: count,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result
}

func (m *PlaylistManager) countSongs(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count
}

// GetPlaylistSongs loads all songs from a playlist.
func (m *PlaylistManager) GetPlaylistSongs(name string) []PlaylistSong {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var songs []PlaylistSong
	pPath := m.playlistPath(name)
	f, err := os.Open(pPath)
	if err != nil {
		return songs
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			song := PlaylistSong{
				Title: parts[0],
				URL:   parts[1],
			}
			if len(parts) >= 3 {
				song.Duration = parts[2]
			}
			songs = append(songs, song)
		}
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

// DeletePlaylist deletes a playlist file.
func (m *PlaylistManager) DeletePlaylist(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return os.Remove(m.playlistPath(name))
}

// RenamePlaylist renames a playlist file.
func (m *PlaylistManager) RenamePlaylist(oldName, newName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldPath := m.playlistPath(oldName)
	newPath := m.playlistPath(newName)
	return os.Rename(oldPath, newPath)
}

// AddSongToPlaylist appends a song to a playlist.
func (m *PlaylistManager) AddSongToPlaylist(playlistName string, song PlaylistSong) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pPath := m.playlistPath(playlistName)
	f, err := os.OpenFile(pPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s|%s|%s\n", song.Title, song.URL, song.Duration)
	_, err = f.WriteString(line)
	return err
}

// RemoveSongFromPlaylist deletes a song at a given index.
func (m *PlaylistManager) RemoveSongFromPlaylist(playlistName string, index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pPath := m.playlistPath(playlistName)
	f, err := os.Open(pPath)
	if err != nil {
		return err
	}
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) != "" {
			lines = append(lines, text)
		}
	}
	f.Close()

	if index < 0 || index >= len(lines) {
		return fmt.Errorf("index out of range")
	}

	lines = append(lines[:index], lines[index+1:]...)

	out, err := os.Create(pPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, l := range lines {
		out.WriteString(l + "\n")
	}
	return nil
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
