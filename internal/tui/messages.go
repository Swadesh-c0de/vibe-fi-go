package tui

import (
	"time"
	"vibe-fi/internal/integration/mpris"
	"vibe-fi/internal/service/lyrics"
	"vibe-fi/internal/service/search"
)

// TickMsg fires at ~60 FPS (16ms) for real-time visualizer physics and audio polling.
type TickMsg time.Time

// RestoreSessionMsg requests restoring the last saved playback session.
type RestoreSessionMsg struct{}

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
type MprisAction = mpris.Action

const (
	MprisNone      = mpris.ActionNone
	MprisPlayPause = mpris.ActionPlayPause
	MprisPlay      = mpris.ActionPlay
	MprisPause     = mpris.ActionPause
	MprisNext      = mpris.ActionNext
	MprisPrevious  = mpris.ActionPrevious
	MprisStop      = mpris.ActionStop
)

// MprisActionMsg delivers a media key action to the Bubble Tea program.
type MprisActionMsg = mpris.ActionMsg
