package search

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
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
