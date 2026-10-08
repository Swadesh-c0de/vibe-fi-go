package mpris

import (
	"testing"
	"vibe-fi/internal/player"
)

func TestPlayerInterfaceMetadata(t *testing.T) {
	mock := player.NewMockPlayer()
	_ = mock.Load("test.mp3", "replace")
	_ = mock.SetProperty("force-media-title", "The Beatles - Hey Jude")

	srv := &Server{
		player: mock,
	}
	pIface := &PlayerInterface{srv: srv}

	meta, err := pIface.Metadata()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	titleVar, ok := meta["xesam:title"]
	if !ok {
		t.Errorf("missing xesam:title in metadata")
	}
	if titleVar.Value() != "Hey Jude" {
		t.Errorf("expected title 'Hey Jude', got '%v'", titleVar.Value())
	}

	artistVar, ok := meta["xesam:artist"]
	if !ok {
		t.Errorf("missing xesam:artist in metadata")
	}
	artists, ok := artistVar.Value().([]string)
	if !ok || len(artists) == 0 || artists[0] != "The Beatles" {
		t.Errorf("expected artist 'The Beatles', got '%v'", artistVar.Value())
	}
}

func TestServerEmitNilSafety(t *testing.T) {
	// Ensure nil receiver or nil props does not panic
	var srv *Server
	srv.EmitPlaybackStatus("Playing")
	srv.EmitTrack("Title", "Artist", 120.0)
	srv.EmitVolume(80)
	srv.EmitSeeked(5000000)
	srv.Stop()

	srv2 := &Server{}
	srv2.EmitPlaybackStatus("Playing")
	srv2.EmitTrack("Title", "Artist", 120.0)
	srv2.EmitVolume(80)
	srv2.EmitSeeked(5000000)
	srv2.Stop()
}
