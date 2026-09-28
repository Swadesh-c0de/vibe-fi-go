package mpris

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"vibe-fi/internal/config"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui"
)

// Server coordinates the MPRIS D-Bus interface.
type Server struct {
	mu      sync.Mutex
	conn    *dbus.Conn
	program *tea.Program
	player  player.AudioPlayer
}

// StartServer starts the D-Bus MPRIS daemon on Linux.
func StartServer(p player.AudioPlayer, prog *tea.Program) *Server {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil
	}

	reply, err := conn.RequestName("org.mpris.MediaPlayer2.vibe_fi", dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		_ = conn.Close()
		return nil
	}

	srv := &Server{
		conn:    conn,
		program: prog,
		player:  p,
	}

	root := &RootInterface{srv: srv}
	playerIface := &PlayerInterface{srv: srv}

	_ = conn.Export(root, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")
	_ = conn.Export(playerIface, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")

	return srv
}

// Stop cleanly unregisters D-Bus name and closes connection.
func (s *Server) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		_, _ = s.conn.ReleaseName("org.mpris.MediaPlayer2.vibe_fi")
		_ = s.conn.Close()
		s.conn = nil
	}
}

// RootInterface implements org.mpris.MediaPlayer2.
type RootInterface struct {
	srv *Server
}

func (r *RootInterface) Raise() *dbus.Error {
	return nil
}

func (r *RootInterface) Quit() *dbus.Error {
	if r.srv.program != nil {
		r.srv.program.Send(tea.Quit())
	}
	return nil
}

func (r *RootInterface) CanQuit() (bool, *dbus.Error)           { return true, nil }
func (r *RootInterface) Fullscreen() (bool, *dbus.Error)        { return false, nil }
func (r *RootInterface) SetFullscreen(bool) *dbus.Error         { return nil }
func (r *RootInterface) CanSetFullscreen() (bool, *dbus.Error)  { return false, nil }
func (r *RootInterface) CanRaise() (bool, *dbus.Error)          { return false, nil }
func (r *RootInterface) HasTrackList() (bool, *dbus.Error)      { return false, nil }
func (r *RootInterface) Identity() (string, *dbus.Error)        { return "Vibe-Fi (" + config.Version + ")", nil }
func (r *RootInterface) SupportedUriSchemes() ([]string, *dbus.Error) {
	return []string{"file", "http", "https"}, nil
}
func (r *RootInterface) SupportedMimeTypes() ([]string, *dbus.Error) {
	return []string{"audio/mpeg", "audio/flac", "audio/ogg", "audio/wav"}, nil
}

// PlayerInterface implements org.mpris.MediaPlayer2.Player.
type PlayerInterface struct {
	srv *Server
}

func (p *PlayerInterface) Next() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(tui.MprisActionMsg{Action: tui.MprisNext})
	}
	return nil
}

func (p *PlayerInterface) Previous() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(tui.MprisActionMsg{Action: tui.MprisPrevious})
	}
	return nil
}

func (p *PlayerInterface) Pause() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(tui.MprisActionMsg{Action: tui.MprisPause})
	}
	return nil
}

func (p *PlayerInterface) PlayPause() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(tui.MprisActionMsg{Action: tui.MprisPlayPause})
	}
	return nil
}

func (p *PlayerInterface) Stop() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(tui.MprisActionMsg{Action: tui.MprisStop})
	}
	return nil
}

func (p *PlayerInterface) Play() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(tui.MprisActionMsg{Action: tui.MprisPlay})
	}
	return nil
}

func (p *PlayerInterface) Seek(offsetMicrosec int64) *dbus.Error {
	sec := float64(offsetMicrosec) / 1000000.0
	_ = p.srv.player.Seek(sec)
	return nil
}

func (p *PlayerInterface) SetPosition(trackID dbus.ObjectPath, positionMicrosec int64) *dbus.Error {
	return nil
}

func (p *PlayerInterface) OpenUri(uri string) *dbus.Error {
	return nil
}

func (p *PlayerInterface) PlaybackStatus() (string, *dbus.Error) {
	if p.srv.player.IsPlaying() {
		return "Playing", nil
	} else if p.srv.player.IsPaused() {
		return "Paused", nil
	}
	return "Stopped", nil
}

func (p *PlayerInterface) Metadata() (map[string]dbus.Variant, *dbus.Error) {
	title := p.srv.player.GetMetadata("media-title")
	if title == "" {
		title = p.srv.player.GetMetadata("filename")
	}
	durMicrosec := int64(p.srv.player.Duration() * 1000000.0)

	meta := map[string]dbus.Variant{
		"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath("/org/mpris/MediaPlayer2/Track/1")),
		"xesam:title":   dbus.MakeVariant(title),
		"mpris:length":  dbus.MakeVariant(durMicrosec),
	}
	return meta, nil
}

func (p *PlayerInterface) Volume() (float64, *dbus.Error) {
	vol := float64(p.srv.player.Volume()) / 100.0
	return vol, nil
}

func (p *PlayerInterface) SetVolume(v float64) *dbus.Error {
	_ = p.srv.player.SetVolume(int(v * 100.0))
	return nil
}

func (p *PlayerInterface) Position() (int64, *dbus.Error) {
	pos := int64(p.srv.player.Position() * 1000000.0)
	return pos, nil
}

func (p *PlayerInterface) CanControl() (bool, *dbus.Error)     { return true, nil }
func (p *PlayerInterface) CanPlay() (bool, *dbus.Error)        { return true, nil }
func (p *PlayerInterface) CanPause() (bool, *dbus.Error)       { return true, nil }
func (p *PlayerInterface) CanSeek() (bool, *dbus.Error)        { return true, nil }
func (p *PlayerInterface) CanGoNext() (bool, *dbus.Error)      { return true, nil }
func (p *PlayerInterface) CanGoPrevious() (bool, *dbus.Error)  { return true, nil }
