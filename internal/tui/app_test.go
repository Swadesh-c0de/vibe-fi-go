package tui

import (
	"testing"

	"vibe-fi/internal/eventbus"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/components"
)

func TestAppModelEventBus(t *testing.T) {
	mock := player.NewMockPlayer()
	app := NewAppModel(mock)

	if app.EventBus == nil {
		t.Fatalf("expected EventBus to be initialized on AppModel")
	}

	var startedEvent *eventbus.TrackStartedEvent
	app.EventBus.Subscribe(eventbus.EventTrackStarted, func(e interface{}) {
		if evt, ok := e.(eventbus.TrackStartedEvent); ok {
			startedEvent = &evt
		}
	})

	var playbackState eventbus.PlaybackState
	app.EventBus.Subscribe(eventbus.EventPlaybackStateChanged, func(e interface{}) {
		if evt, ok := e.(eventbus.PlaybackStateChangedEvent); ok {
			playbackState = evt.State
		}
	})

	_ = app.StartTrackPlayback("Chill Beats", "/path/to/chill.mp3", "02:30", "Artist A")

	if startedEvent == nil {
		t.Fatalf("expected EventTrackStarted to be published")
	}
	if startedEvent.Title != "Chill Beats" {
		t.Errorf("expected Title 'Chill Beats', got %q", startedEvent.Title)
	}
	if startedEvent.URL != "/path/to/chill.mp3" {
		t.Errorf("expected URL '/path/to/chill.mp3', got %q", startedEvent.URL)
	}
	if playbackState != eventbus.StatePlaying {
		t.Errorf("expected StatePlaying, got %v", playbackState)
	}
}

func TestAppModelViewRendering(t *testing.T) {
	mock := player.NewMockPlayer()
	app := NewAppModel(mock)
	app.Width = 80
	app.Height = 24

	// 1. Intro View
	app.SetMode(components.ViewModeIntro)
	v := app.View()
	if v == "" {
		t.Errorf("expected non-empty View in Intro mode")
	}

	// 2. Playback View
	app.SetMode(components.ViewModePlayback)
	v = app.View()
	if v == "" {
		t.Errorf("expected non-empty View in Playback mode")
	}

	// 3. Help Modal
	app.ShowHelpModal = true
	v = app.View()
	if v == "" {
		t.Errorf("expected non-empty View with Help Modal")
	}
}
