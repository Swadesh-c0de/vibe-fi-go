package search

import (
	"testing"
	"time"
)

func TestStreamCacheGetSet(t *testing.T) {
	cache := NewStreamCache()

	url := "https://www.youtube.com/watch?v=test1234"
	info := StreamInfo{
		Title:     "Test Song",
		Artist:    "Test Artist",
		Duration:  180,
		StreamURL: "https://rr1---sn-test.googlevideo.com/videoplayback",
	}

	if _, ok := cache.Get(url); ok {
		t.Fatal("expected empty cache on init")
	}

	cache.Set(url, info)

	got, ok := cache.Get(url)
	if !ok {
		t.Fatal("expected item to be in cache")
	}
	if got.Title != info.Title || got.StreamURL != info.StreamURL {
		t.Errorf("mismatched cached info: got %+v, want %+v", got, info)
	}

	if !cache.Has(url) {
		t.Error("expected Has(url) to be true")
	}

	// Test expiration
	cache.TTL = 10 * time.Millisecond
	time.Sleep(15 * time.Millisecond)

	if _, ok := cache.Get(url); ok {
		t.Error("expected entry to be expired")
	}

	// Test clear
	cache.Set(url, info)
	cache.Clear()
	if _, ok := cache.Get(url); ok {
		t.Error("expected empty cache after Clear()")
	}
}
