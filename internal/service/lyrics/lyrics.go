package lyrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"vibe-fi/internal/config"
	"vibe-fi/internal/utils/net"
	"vibe-fi/internal/utils/stringutil"
)

// LyricLine represents a single line of time-synced lyrics.
type LyricLine struct {
	Timestamp float64 `json:"timestamp"` // in seconds
	Text      string  `json:"text"`
}

// LyricsData holds lyrics and parsed synced lines.
type LyricsData struct {
	PlainLyrics  string      `json:"plain_lyrics"`
	SyncedLyrics []LyricLine `json:"synced_lyrics"`
	HasSynced    bool        `json:"has_synced"`
}

type lrclibResponse struct {
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// LyricsManager handles lyrics retrieval, parsing, and caching.
type LyricsManager struct {
	client   *http.Client
	cacheDir string
}

// NewLyricsManager creates a new LyricsManager instance.
func NewLyricsManager() *LyricsManager {
	return &LyricsManager{
		client:   &http.Client{Timeout: 8 * time.Second},
		cacheDir: config.GetLyricsCacheDir(),
	}
}

func (m *LyricsManager) cacheKey(artist, title string) string {
	combined := fmt.Sprintf("%s_%s", strings.ToLower(artist), strings.ToLower(title))
	reg := regexp.MustCompile(`[^a-z0-9_-]+`)
	key := reg.ReplaceAllString(combined, "+")
	return filepath.Join(m.cacheDir, key+".json")
}

// GetCachedLyrics looks up lyrics in the local disk cache.
func (m *LyricsManager) GetCachedLyrics(artist, title string) (LyricsData, bool) {
	var data LyricsData
	cacheFile := m.cacheKey(artist, title)
	f, err := os.Open(cacheFile)
	if err != nil {
		return data, false
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&data); err == nil {
		return data, true
	}
	return data, false
}

func (m *LyricsManager) saveToCache(artist, title string, data LyricsData) {
	cacheFile := m.cacheKey(artist, title)
	f, err := os.Create(cacheFile)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(data)
}

var timestampRegex = regexp.MustCompile(`^\[(\d{2}):(\d{2}(?:\.\d+)?)\](.*)$`)

// ParseSyncedLyrics parses raw LRC formatted string into LyricLine slice.
func ParseSyncedLyrics(lrc string) []LyricLine {
	var lines []LyricLine
	scanner := strings.Split(lrc, "\n")

	for _, rawLine := range scanner {
		rawLine = strings.TrimSpace(rawLine)
		if rawLine == "" {
			continue
		}
		matches := timestampRegex.FindStringSubmatch(rawLine)
		if len(matches) == 4 {
			mins, _ := strconv.Atoi(matches[1])
			secs, _ := strconv.ParseFloat(matches[2], 64)
			text := strings.TrimSpace(matches[3])

			lines = append(lines, LyricLine{
				Timestamp: float64(mins)*60.0 + secs,
				Text:      text,
			})
		}
	}

	return lines
}

// FetchLyrics searches lrclib.net for lyrics matching artist and title.
func (m *LyricsManager) FetchLyrics(artist, title string, duration float64) (LyricsData, error) {
	if artist == "" && title != "" {
		artist, title = stringutil.CleanTrackTitle(title)
	}

	// Check offline cache first
	if cached, ok := m.GetCachedLyrics(artist, title); ok {
		return cached, nil
	}

	if !net.IsOnline() {
		return LyricsData{PlainLyrics: "Network unavailable: lyrics offline."}, fmt.Errorf("offline")
	}

	// Step 1: Try exact match /api/get
	data, err := m.fetchExact(artist, title, duration)
	if err == nil && (data.HasSynced || data.PlainLyrics != "") {
		m.saveToCache(artist, title, data)
		return data, nil
	}

	// Step 2: Try search fallback /api/search?q=
	query := title
	if artist != "" {
		query = artist + " " + title
	}
	data, err = m.fetchSearch(query)
	if err == nil && (data.HasSynced || data.PlainLyrics != "") {
		m.saveToCache(artist, title, data)
		return data, nil
	}

	notFound := LyricsData{PlainLyrics: "Lyrics not found"}
	return notFound, nil
}

func (m *LyricsManager) fetchExact(artist, title string, duration float64) (LyricsData, error) {
	var data LyricsData
	params := url.Values{}
	params.Set("track_name", title)
	if artist != "" {
		params.Set("artist_name", artist)
	}
	if duration > 0 {
		params.Set("duration", strconv.Itoa(int(duration)))
	}

	reqURL := "https://lrclib.net/api/get?" + params.Encode()
	resp, err := m.client.Get(reqURL)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	var res lrclibResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return data, err
	}

	if res.SyncedLyrics != "" {
		data.SyncedLyrics = ParseSyncedLyrics(res.SyncedLyrics)
		data.HasSynced = len(data.SyncedLyrics) > 0
	}
	data.PlainLyrics = res.PlainLyrics
	return data, nil
}

func (m *LyricsManager) fetchSearch(query string) (LyricsData, error) {
	var data LyricsData
	reqURL := "https://lrclib.net/api/search?q=" + url.QueryEscape(query)
	resp, err := m.client.Get(reqURL)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	var candidates []lrclibResponse
	if err := json.Unmarshal(body, &candidates); err != nil || len(candidates) == 0 {
		return data, fmt.Errorf("no results")
	}

	// Pick first candidate with synced lyrics or first candidate with plain lyrics
	var chosen lrclibResponse
	for _, c := range candidates {
		if c.SyncedLyrics != "" {
			chosen = c
			break
		}
	}
	if chosen.SyncedLyrics == "" && len(candidates) > 0 {
		chosen = candidates[0]
	}

	if chosen.SyncedLyrics != "" {
		data.SyncedLyrics = ParseSyncedLyrics(chosen.SyncedLyrics)
		data.HasSynced = len(data.SyncedLyrics) > 0
	}
	data.PlainLyrics = chosen.PlainLyrics
	return data, nil
}
