package player

// AudioLevelStats contains real-time audio analysis from FFmpeg @astats filter.
type AudioLevelStats struct {
	RMSOverall    float32 // Normalized perceived loudness (0.0 to 1.0)
	PeakOverall   float32 // Normalized instantaneous peak hit (0.0 to 1.0)
	RMSLeft       float32 // Left channel loudness (0.0 to 1.0)
	RMSRight      float32 // Right channel loudness (0.0 to 1.0)
	ZeroCrossings float32 // Pitch / tone estimator (low = deep bass, high = bright/harsh)
	Valid         bool
}

// EventType defines audio player lifecycle events.
type EventType int

const (
	EventNone EventType = iota
	EventStartFile
	EventFileLoaded
	EventEndFileEOF
	EventEndFileError
	EventEndFileStop
)

// PlayerEvent represents an event received from mpv.
type PlayerEvent struct {
	Type  EventType
	Error string
}

// AudioPlayer defines the interface for audio playback and analysis.
type AudioPlayer interface {
	Load(path string, mode string) error
	Play() error
	Pause() error
	TogglePause() error
	Stop() error
	Seek(seconds float64) error
	IsPlaying() bool
	IsPaused() bool
	IsIdle() bool
	IsLoading() bool
	IsBuffering() bool
	Position() float64
	Duration() float64
	Volume() int
	SetVolume(volume int) error
	GetMetadata(key string) string
	SetProperty(name, value string) error
	GetAudioStats() AudioLevelStats
	PollEvents() []PlayerEvent
	HasTrackFinished() bool
	ConsumeTrackFinished() bool
	HasPlaybackError() bool
	ConsumePlaybackError() bool
	GetLastError() string
	ClearPlaybackFlags()
	Close() error
}
