package player

import "sync"

// MockPlayer implements AudioPlayer for headless environments and tests.
type MockPlayer struct {
	mu            sync.Mutex
	path          string
	playing       bool
	paused        bool
	idle          bool
	loading       bool
	buffering     bool
	pos           float64
	dur           float64
	volume        int
	metadata      map[string]string
	trackFinished bool
	playbackError bool
	lastError     string
}

// NewMockPlayer creates a MockPlayer instance.
func NewMockPlayer() *MockPlayer {
	return &MockPlayer{
		idle:     true,
		volume:   100,
		metadata: make(map[string]string),
	}
}

func (m *MockPlayer) Load(path string, mode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.path = path
	m.idle = false
	m.playing = true
	m.paused = false
	m.loading = false
	m.pos = 0
	m.dur = 180.0
	m.trackFinished = false
	m.playbackError = false
	return nil
}

func (m *MockPlayer) Play() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playing = true
	m.paused = false
	m.idle = false
	return nil
}

func (m *MockPlayer) Pause() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused = true
	m.playing = false
	return nil
}

func (m *MockPlayer) TogglePause() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused = !m.paused
	m.playing = !m.paused
	return nil
}

func (m *MockPlayer) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playing = false
	m.paused = false
	m.idle = true
	m.pos = 0
	return nil
}

func (m *MockPlayer) Seek(seconds float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pos += seconds
	if m.pos < 0 {
		m.pos = 0
	}
	return nil
}

func (m *MockPlayer) IsPlaying() bool   { m.mu.Lock(); defer m.mu.Unlock(); return m.playing }
func (m *MockPlayer) IsPaused() bool    { m.mu.Lock(); defer m.mu.Unlock(); return m.paused }
func (m *MockPlayer) IsIdle() bool      { m.mu.Lock(); defer m.mu.Unlock(); return m.idle }
func (m *MockPlayer) IsLoading() bool   { m.mu.Lock(); defer m.mu.Unlock(); return m.loading }
func (m *MockPlayer) IsBuffering() bool { m.mu.Lock(); defer m.mu.Unlock(); return m.buffering }
func (m *MockPlayer) Position() float64 { m.mu.Lock(); defer m.mu.Unlock(); return m.pos }
func (m *MockPlayer) Duration() float64 { m.mu.Lock(); defer m.mu.Unlock(); return m.dur }
func (m *MockPlayer) Volume() int       { m.mu.Lock(); defer m.mu.Unlock(); return m.volume }

func (m *MockPlayer) SetVolume(volume int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.volume = volume
	return nil
}

func (m *MockPlayer) GetMetadata(key string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metadata[key]
}

func (m *MockPlayer) SetProperty(name, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[name] = value
	return nil
}

func (m *MockPlayer) GetAudioStats() AudioLevelStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.playing {
		return AudioLevelStats{ZeroCrossings: 0.05}
	}
	return AudioLevelStats{
		RMSOverall:    0.45,
		PeakOverall:   0.65,
		RMSLeft:       0.42,
		RMSRight:      0.48,
		ZeroCrossings: 0.25,
		Valid:         true,
	}
}

func (m *MockPlayer) PollEvents() []PlayerEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil
}

func (m *MockPlayer) HasTrackFinished() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.trackFinished
}

func (m *MockPlayer) ConsumeTrackFinished() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.trackFinished {
		m.trackFinished = false
		return true
	}
	return false
}

func (m *MockPlayer) HasPlaybackError() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playbackError
}

func (m *MockPlayer) ConsumePlaybackError() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.playbackError {
		m.playbackError = false
		return true
	}
	return false
}

func (m *MockPlayer) GetLastError() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastError
}

func (m *MockPlayer) ClearPlaybackFlags() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trackFinished = false
	m.playbackError = false
	m.loading = false
	m.lastError = ""
}

func (m *MockPlayer) Close() error {
	return nil
}
