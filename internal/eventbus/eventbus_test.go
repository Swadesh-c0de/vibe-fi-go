package eventbus

import (
	"sync"
	"testing"
)

func TestEventBusPublishSubscribe(t *testing.T) {
	bus := New()

	var receivedTitle string
	var receivedArtist string
	bus.Subscribe(EventTrackStarted, func(e interface{}) {
		evt, ok := e.(TrackStartedEvent)
		if !ok {
			t.Errorf("expected TrackStartedEvent, got %T", e)
			return
		}
		receivedTitle = evt.Title
		receivedArtist = evt.Artist
	})

	var receivedVol int
	bus.Subscribe(EventVolumeChanged, func(e interface{}) {
		if evt, ok := e.(VolumeChangedEvent); ok {
			receivedVol = evt.Volume
		}
	})

	bus.Publish(EventTrackStarted, TrackStartedEvent{
		Title:  "Track Title",
		Artist: "Artist Name",
	})

	if receivedTitle != "Track Title" || receivedArtist != "Artist Name" {
		t.Errorf("TrackStarted not handled properly: got %s - %s", receivedArtist, receivedTitle)
	}

	bus.Publish(EventVolumeChanged, VolumeChangedEvent{Volume: 85})
	if receivedVol != 85 {
		t.Errorf("expected volume 85, got %d", receivedVol)
	}
}

func TestEventBusConcurrent(t *testing.T) {
	bus := New()
	var wg sync.WaitGroup

	var count int
	var mu sync.Mutex

	bus.Subscribe(EventPlaybackStateChanged, func(e interface{}) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(EventPlaybackStateChanged, PlaybackStateChangedEvent{State: StatePlaying})
		}()
	}

	wg.Wait()

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != 50 {
		t.Errorf("expected 50 events received concurrently, got %d", finalCount)
	}
}
