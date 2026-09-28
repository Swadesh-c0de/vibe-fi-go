package tui

import (
	"time"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/service/search"
)

// TickMsg fires at ~30 FPS for visualizer and audio state polling.
type TickMsg time.Time

// StatusMsg displays a temporary message in the help bar for 3 seconds.
type StatusMsg struct {
	Message string
}

// ClearStatusMsg removes the transient message when it expires.
type ClearStatusMsg struct{}

// LyricsFetchedMsg delivers asynchronously fetched lyrics.
type LyricsFetchedMsg struct {
	RequestID uint64
	Title     string
	Artist    string
	Data      lyrics.LyricsData
	Err       error
}

// SearchResultsMsg delivers async YouTube search results.
type SearchResultsMsg struct {
	Query   string
	Results []search.SearchResult
	Err     error
}

// StreamResolvedMsg delivers async YouTube URL stream resolution.
type StreamResolvedMsg struct {
	URL  string
	Info search.StreamInfo
	Err  error
}

// UpdateDiscoveredMsg notifies when a newer version is available.
type UpdateDiscoveredMsg struct {
	Version string
}

// MprisAction defines media key actions from Linux D-Bus.
type MprisAction int

const (
	MprisNone MprisAction = iota
	MprisPlayPause
	MprisPlay
	MprisPause
	MprisNext
	MprisPrevious
	MprisStop
)

// MprisActionMsg delivers a media key action to the Bubble Tea program.
type MprisActionMsg struct {
	Action MprisAction
}
