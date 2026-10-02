package mpris

import (
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"vibe-fi/internal/config"
	"vibe-fi/internal/player"
)

// Action defines media key actions from Linux D-Bus.
type Action int

const (
	ActionNone Action = iota
	ActionPlayPause
	ActionPlay
	ActionPause
	ActionNext
	ActionPrevious
	ActionStop
)

// ActionMsg delivers an MPRIS media key action to the Bubble Tea program.
type ActionMsg struct {
	Action Action
}

// Server coordinates the MPRIS D-Bus interface.
type Server struct {
	mu      sync.Mutex
	conn    *dbus.Conn
	program *tea.Program
	player  player.AudioPlayer
	props   *prop.Properties
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
	_ = conn.ExportWithMap(playerIface, map[string]string{
		"MprisSeek": "Seek",
	}, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")

	// Export standard org.freedesktop.DBus.Properties interface
	propsMap := prop.Map{
		"org.mpris.MediaPlayer2": {
			"CanQuit":             {Value: true, Emit: prop.EmitConst},
			"Fullscreen":          {Value: false, Emit: prop.EmitFalse},
			"CanSetFullscreen":    {Value: false, Emit: prop.EmitConst},
			"CanRaise":            {Value: false, Emit: prop.EmitConst},
			"HasTrackList":        {Value: false, Emit: prop.EmitConst},
			"Identity":            {Value: "Vibe-Fi (" + config.Version + ")", Emit: prop.EmitConst},
			"SupportedUriSchemes": {Value: []string{"file", "http", "https"}, Emit: prop.EmitConst},
			"SupportedMimeTypes":  {Value: []string{"audio/mpeg", "audio/flac", "audio/ogg", "audio/wav"}, Emit: prop.EmitConst},
		},
		"org.mpris.MediaPlayer2.Player": {
			"PlaybackStatus": {Value: "Stopped", Emit: prop.EmitTrue},
			"LoopStatus":     {Value: "None", Emit: prop.EmitTrue},
			"Rate":           {Value: 1.0, Emit: prop.EmitFalse},
			"Shuffle":        {Value: false, Emit: prop.EmitTrue},
			"Metadata":       {Value: map[string]dbus.Variant{}, Emit: prop.EmitTrue},
			"Volume": {
				Value:    1.0,
				Writable: true,
				Emit:     prop.EmitTrue,
				Callback: func(c *prop.Change) *dbus.Error {
					if v, ok := c.Value.(float64); ok {
						_ = p.SetVolume(int(v * 100.0))
					}
					return nil
				},
			},
			"Position":      {Value: int64(0), Emit: prop.EmitFalse},
			"MinimumRate":   {Value: 1.0, Emit: prop.EmitConst},
			"MaximumRate":   {Value: 1.0, Emit: prop.EmitConst},
			"CanControl":    {Value: true, Emit: prop.EmitConst},
			"CanPlay":       {Value: true, Emit: prop.EmitConst},
			"CanPause":      {Value: true, Emit: prop.EmitConst},
			"CanSeek":       {Value: true, Emit: prop.EmitConst},
			"CanGoNext":     {Value: true, Emit: prop.EmitConst},
			"CanGoPrevious": {Value: true, Emit: prop.EmitConst},
		},
	}

	props, err := prop.Export(conn, "/org/mpris/MediaPlayer2", propsMap)
	if err == nil {
		srv.props = props
	}

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

// EmitPlaybackStatus updates the PlaybackStatus property and sends PropertiesChanged signal.
func (s *Server) EmitPlaybackStatus(status string) {
	if s == nil || s.props == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.props.Set("org.mpris.MediaPlayer2.Player", "PlaybackStatus", dbus.MakeVariant(status))
}

// EmitTrack updates Metadata property and emits PropertiesChanged signal to desktop notifications.
func (s *Server) EmitTrack(title, artist string, durSec float64) {
	if s == nil || s.props == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if artist == "" && strings.Contains(title, " - ") {
		parts := strings.SplitN(title, " - ", 2)
		artist = parts[0]
		title = parts[1]
	}

	durMicrosec := int64(durSec * 1000000.0)
	meta := map[string]dbus.Variant{
		"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath("/org/mpris/MediaPlayer2/Track/1")),
		"xesam:title":   dbus.MakeVariant(title),
		"mpris:length":  dbus.MakeVariant(durMicrosec),
	}
	if artist != "" {
		meta["xesam:artist"] = dbus.MakeVariant([]string{artist})
	}

	_ = s.props.Set("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(meta))
}

// EmitVolume updates Volume property and emits PropertiesChanged signal.
func (s *Server) EmitVolume(volPercent int) {
	if s == nil || s.props == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	vol := float64(volPercent) / 100.0
	_ = s.props.Set("org.mpris.MediaPlayer2.Player", "Volume", dbus.MakeVariant(vol))
}

// EmitSeeked emits the org.mpris.MediaPlayer2.Player.Seeked signal.
func (s *Server) EmitSeeked(posMicrosec int64) {
	if s == nil || s.conn == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.conn.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player.Seeked", posMicrosec)
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
		p.srv.program.Send(ActionMsg{Action: ActionNext})
	}
	return nil
}

func (p *PlayerInterface) Previous() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(ActionMsg{Action: ActionPrevious})
	}
	return nil
}

func (p *PlayerInterface) Pause() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(ActionMsg{Action: ActionPause})
		p.srv.EmitPlaybackStatus("Paused")
	}
	return nil
}

func (p *PlayerInterface) PlayPause() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(ActionMsg{Action: ActionPlayPause})
	}
	return nil
}

func (p *PlayerInterface) Stop() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(ActionMsg{Action: ActionStop})
		p.srv.EmitPlaybackStatus("Stopped")
	}
	return nil
}

func (p *PlayerInterface) Play() *dbus.Error {
	if p.srv.program != nil {
		p.srv.program.Send(ActionMsg{Action: ActionPlay})
		p.srv.EmitPlaybackStatus("Playing")
	}
	return nil
}

func (p *PlayerInterface) MprisSeek(offsetMicrosec int64) *dbus.Error {
	if p.srv == nil || p.srv.player == nil {
		return nil
	}
	sec := float64(offsetMicrosec) / 1000000.0
	_ = p.srv.player.Seek(sec)
	newPos := int64(p.srv.player.Position() * 1000000.0)
	p.srv.EmitSeeked(newPos)
	return nil
}

func (p *PlayerInterface) SetPosition(trackID dbus.ObjectPath, positionMicrosec int64) *dbus.Error {
	if p.srv == nil || p.srv.player == nil {
		return nil
	}
	targetSec := float64(positionMicrosec) / 1000000.0
	delta := targetSec - p.srv.player.Position()
	_ = p.srv.player.Seek(delta)
	newPos := int64(p.srv.player.Position() * 1000000.0)
	p.srv.EmitSeeked(newPos)
	return nil
}

func (p *PlayerInterface) OpenUri(uri string) *dbus.Error {
	return nil
}

func (p *PlayerInterface) PlaybackStatus() (string, *dbus.Error) {
	if p.srv == nil || p.srv.player == nil {
		return "Stopped", nil
	}
	if p.srv.player.IsPlaying() {
		return "Playing", nil
	} else if p.srv.player.IsPaused() {
		return "Paused", nil
	}
	return "Stopped", nil
}

func (p *PlayerInterface) Metadata() (map[string]dbus.Variant, *dbus.Error) {
	if p.srv == nil || p.srv.player == nil {
		return map[string]dbus.Variant{}, nil
	}
	title := p.srv.player.GetMetadata("media-title")
	if title == "" {
		title = p.srv.player.GetMetadata("force-media-title")
	}
	if title == "" {
		title = p.srv.player.GetMetadata("filename")
	}
	artist := p.srv.player.GetMetadata("artist")
	if artist == "" && strings.Contains(title, " - ") {
		parts := strings.SplitN(title, " - ", 2)
		artist = parts[0]
		title = parts[1]
	}
	durMicrosec := int64(p.srv.player.Duration() * 1000000.0)

	meta := map[string]dbus.Variant{
		"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath("/org/mpris/MediaPlayer2/Track/1")),
		"xesam:title":   dbus.MakeVariant(title),
		"mpris:length":  dbus.MakeVariant(durMicrosec),
	}
	if artist != "" {
		meta["xesam:artist"] = dbus.MakeVariant([]string{artist})
	}
	return meta, nil
}

func (p *PlayerInterface) Volume() (float64, *dbus.Error) {
	if p.srv == nil || p.srv.player == nil {
		return 1.0, nil
	}
	vol := float64(p.srv.player.Volume()) / 100.0
	return vol, nil
}

func (p *PlayerInterface) SetVolume(v float64) *dbus.Error {
	if p.srv == nil || p.srv.player == nil {
		return nil
	}
	volPct := int(v * 100.0)
	_ = p.srv.player.SetVolume(volPct)
	p.srv.EmitVolume(volPct)
	return nil
}

func (p *PlayerInterface) Position() (int64, *dbus.Error) {
	if p.srv == nil || p.srv.player == nil {
		return 0, nil
	}
	pos := int64(p.srv.player.Position() * 1000000.0)
	return pos, nil
}

func (p *PlayerInterface) CanControl() (bool, *dbus.Error)     { return true, nil }
func (p *PlayerInterface) CanPlay() (bool, *dbus.Error)        { return true, nil }
func (p *PlayerInterface) CanPause() (bool, *dbus.Error)       { return true, nil }
func (p *PlayerInterface) CanSeek() (bool, *dbus.Error)        { return true, nil }
func (p *PlayerInterface) CanGoNext() (bool, *dbus.Error)      { return true, nil }
func (p *PlayerInterface) CanGoPrevious() (bool, *dbus.Error)  { return true, nil }
