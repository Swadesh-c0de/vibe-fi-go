package search

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"vibe-fi/internal/utils/bottle"
	"vibe-fi/internal/utils/net"
	"vibe-fi/internal/utils/stringutil"
)

// SearchResult represents a single YouTube search hit.
type SearchResult struct {
	Title    string
	URL      string
	Duration string
	Uploader string
}

// StreamInfo represents resolved direct stream info.
type StreamInfo struct {
	Title     string
	Artist    string
	Duration  float64
	StreamURL string
}

// SearchYouTube invokes yt-dlp safely and parses results.
func SearchYouTube(query string, limit int) ([]SearchResult, error) {
	var results []SearchResult
	query = strings.TrimSpace(query)
	if query == "" {
		return results, nil
	}
	if !net.IsOnline() {
		return results, fmt.Errorf("network unavailable")
	}

	ytdlPath := bottle.FindExecutable("yt-dlp")
	if ytdlPath == "" {
		return results, fmt.Errorf("yt-dlp not found")
	}

	searchTerm := fmt.Sprintf("ytsearch%d:%s", limit, query)
	args := []string{
		"--print",
		"%(title)s|%(uploader)s|%(webpage_url)s|%(duration_string)s",
		"--flat-playlist",
		"--no-warnings",
		"--extractor-args", "youtube:player_client=android,web",
		searchTerm,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ytdlPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return results, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if line == "" {
			continue
		}

		// Parse from right: duration, URL, uploader, title
		lastPipe := strings.LastIndex(line, "|")
		if lastPipe == -1 {
			continue
		}
		secondLast := strings.LastIndex(line[:lastPipe], "|")
		if secondLast == -1 {
			continue
		}
		thirdLast := strings.LastIndex(line[:secondLast], "|")

		var rawTitle, uploader string
		if thirdLast != -1 {
			rawTitle = line[:thirdLast]
			uploader = line[thirdLast+1 : secondLast]
		} else {
			rawTitle = line[:secondLast]
		}

		title := stringutil.SanitizeText(rawTitle)
		url := line[secondLast+1 : lastPipe]
		duration := line[lastPipe+1:]

		if uploader != "" && uploader != "NA" {
			uploader = strings.TrimSuffix(uploader, " - Topic")
			if !strings.Contains(title, " - ") && uploader != "" {
				title = uploader + " - " + title
			}
		}

		if duration == "" || duration == "NA" {
			duration = "--:--"
		}

		if title != "" && url != "" {
			results = append(results, SearchResult{
				Title:    title,
				URL:      url,
				Duration: duration,
				Uploader: uploader,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return results, err
	}

	return results, nil
}

// ResolveStreamInfo extracts media title, artist, duration, and direct stream URL.
func ResolveStreamInfo(url string) (StreamInfo, error) {
	var info StreamInfo
	if !net.IsOnline() {
		return info, fmt.Errorf("network unavailable")
	}

	ytdlPath := bottle.FindExecutable("yt-dlp")
	if ytdlPath == "" {
		return info, fmt.Errorf("yt-dlp not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	args := []string{
		"-f", "251/140/bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio/best",
		"--print", "%(title)s|%(uploader)s|%(duration_string)s",
		"-g",
		"--no-warnings",
		"--extractor-args", "youtube:player_client=android,web",
		url,
	}

	cmd := exec.CommandContext(ctx, ytdlPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return info, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return info, fmt.Errorf("no output from stream resolver")
	}

	// First line is metadata %(title)s|%(uploader)s|%(duration_string)s
	metaLine := strings.TrimRight(lines[0], "\r")
	metaParts := strings.Split(metaLine, "|")
	if len(metaParts) >= 3 {
		info.Title = metaParts[0]
		info.Artist = strings.TrimSuffix(metaParts[1], " - Topic")
		info.Duration = stringutil.ParseDuration(metaParts[2])
	} else if len(metaParts) > 0 {
		info.Title = metaParts[0]
	}

	// Last non-empty line is stream URL
	for i := len(lines) - 1; i >= 1; i-- {
		candidate := strings.TrimSpace(lines[i])
		if candidate != "" && stringutil.IsURL(candidate) {
			info.StreamURL = candidate
			break
		}
	}
	if info.StreamURL == "" && len(lines) >= 2 {
		info.StreamURL = strings.TrimSpace(lines[len(lines)-1])
	}
	if info.StreamURL == "" {
		info.StreamURL = url
	}

	return info, nil
}

type cachedStream struct {
	info      StreamInfo
	timestamp time.Time
}

// StreamCache provides thread-safe in-memory caching and background pre-fetching
// of direct audio streaming URLs (e.g. from yt-dlp) to enable gapless queue playback.
type StreamCache struct {
	mu      sync.RWMutex
	streams map[string]cachedStream
	pending map[string]bool
	TTL     time.Duration
}

// DefaultStreamCache is the global stream cache instance.
var DefaultStreamCache = NewStreamCache()

// NewStreamCache constructs an initialized StreamCache with a 4-hour TTL.
func NewStreamCache() *StreamCache {
	return &StreamCache{
		streams: make(map[string]cachedStream),
		pending: make(map[string]bool),
		TTL:     4 * time.Hour,
	}
}

// Get retrieves a cached StreamInfo if present and not expired.
func (c *StreamCache) Get(url string) (StreamInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.streams[url]
	if !ok {
		return StreamInfo{}, false
	}
	if time.Since(entry.timestamp) > c.TTL {
		return StreamInfo{}, false
	}
	return entry.info, true
}

// Set stores a resolved StreamInfo in the cache.
func (c *StreamCache) Set(url string, info StreamInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.streams[url] = cachedStream{
		info:      info,
		timestamp: time.Now(),
	}
	delete(c.pending, url)
}

// Has checks if a valid cached entry exists.
func (c *StreamCache) Has(url string) bool {
	_, ok := c.Get(url)
	return ok
}

// PreFetch asynchronously resolves and caches stream info for url in the background.
// If url is already cached or currently resolving, PreFetch returns immediately.
func (c *StreamCache) PreFetch(url string) {
	if url == "" {
		return
	}
	c.mu.Lock()
	if entry, ok := c.streams[url]; ok && time.Since(entry.timestamp) <= c.TTL {
		c.mu.Unlock()
		return
	}
	if c.pending[url] {
		c.mu.Unlock()
		return
	}
	c.pending[url] = true
	c.mu.Unlock()

	go func() {
		info, err := ResolveStreamInfo(url)
		if err == nil && info.StreamURL != "" {
			c.Set(url, info)
		} else {
			c.mu.Lock()
			delete(c.pending, url)
			c.mu.Unlock()
		}
	}()
}

// Clear flushes all cached entries.
func (c *StreamCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.streams = make(map[string]cachedStream)
	c.pending = make(map[string]bool)
}
