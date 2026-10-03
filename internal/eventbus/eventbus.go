package eventbus

import "sync"

// EventType identifies the category of domain event.
type EventType int

const (
	EventTrackStarted EventType = iota
	EventPlaybackStateChanged
	EventVolumeChanged
	EventSeeked
	EventTrackEnded
)

// PlaybackState represents playback activity.
type PlaybackState int

const (
	StateStopped PlaybackState = iota
	StatePlaying
	StatePaused
)

func (s PlaybackState) String() string {
	switch s {
	case StatePlaying:
		return "Playing"
	case StatePaused:
		return "Paused"
	default:
		return "Stopped"
	}
}

// TrackStartedEvent is published when a new audio track begins playback.
type TrackStartedEvent struct {
	Title    string
	Artist   string
	URL      string
	Duration float64
	Playlist string
}

// PlaybackStateChangedEvent is published on play, pause, or stop.
type PlaybackStateChangedEvent struct {
	State PlaybackState
}

// VolumeChangedEvent is published on volume adjustments.
type VolumeChangedEvent struct {
	Volume int // 0 to 100
}

// SeekedEvent is published when playback position seeks.
type SeekedEvent struct {
	PositionSeconds float64
}

// TrackEndedEvent is published when a track ends naturally or errors out.
type TrackEndedEvent struct {
	URL   string
	Error error
}

// Handler is a callback invoked when an event of matching type occurs.
type Handler func(event interface{})

// Bus manages thread-safe event subscriptions and publishing.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]Handler
}

// New creates an initialized EventBus.
func New() *Bus {
	return &Bus{
		subscribers: make(map[EventType][]Handler),
	}
}

// Subscribe registers a handler for a given event type.
func (b *Bus) Subscribe(eventType EventType, handler Handler) {
	if b == nil || handler == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish notifies all registered subscribers of an event synchronously.
func (b *Bus) Publish(eventType EventType, event interface{}) {
	if b == nil {
		return
	}
	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[eventType]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}
