package discord

import (
	"testing"
)

func TestDiscordClientOffline(t *testing.T) {
	client := NewClient("1234567890")
	if client == nil {
		t.Fatalf("expected NewClient to return a non-nil client")
	}

	// Should safely no-op when Discord socket is not open
	client.UpdatePresence("Song Title", "Artist Name")
	client.ClearPresence()
	client.Close()
}
