package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"vibe-fi/internal/config"
	"vibe-fi/internal/eventbus"
	"vibe-fi/internal/integration/discord"
	"vibe-fi/internal/integration/mpris"
	"vibe-fi/internal/player"
	"vibe-fi/internal/service/library"
	"vibe-fi/internal/service/playlist"
	"vibe-fi/internal/service/search"
	"vibe-fi/internal/service/updater"
	"vibe-fi/internal/tui"
	"vibe-fi/internal/tui/components"
	"vibe-fi/internal/utils/bottle"
	"vibe-fi/internal/utils/mem"
	"vibe-fi/internal/utils/net"
	"vibe-fi/internal/utils/stringutil"
)

func printHelp(progName string) {
	fmt.Printf("Usage: %s [OPTIONS] [FILE | URL | QUERY]\n\n", progName)
	fmt.Println("Vibe-Fi v2: Free, open-source terminal music player for Linux & macOS.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Printf("  %s --help | -h          Show this help message\n", progName)
	fmt.Printf("  %s --version | -v       Show version information\n", progName)
	fmt.Printf("  %s --bottle | -b        Inspect bottle dependency status\n", progName)
	fmt.Printf("  %s --restore | -r       Restore last session track, seek & volume\n", progName)
	fmt.Printf("  %s --no-update          Skip startup update check\n", progName)
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  SPACE       Play / Pause toggle")
	fmt.Println("  N / B       Next / Previous track in queue")
	fmt.Println("  Left/Right  Seek backward / forward 5s")
	fmt.Println("  +/-         Volume adjustment")
	fmt.Println("  S           Search YouTube")
	fmt.Println("  L           Browse Local Music Library")
	fmt.Println("  P           Manage Playlists")
	fmt.Println("  C           Interactive Play Queue")
	fmt.Println("  T           Cycle Themes (Midnight, Nord, Matrix, HyDE, Gruvbox, Slate)")
	fmt.Println("  ESC / Q     Back / Quit")
}

func printBottleStatus() {
	fmt.Printf(":: Vibe-Fi Bottle Directory: %s\n", config.GetBottleDir())
	fmt.Printf(":: Bottle Bin: %s\n", config.GetBottleBinDir())
	ytdl := bottle.FindExecutable("yt-dlp")
	if ytdl != "" {
		fmt.Printf(":: Found yt-dlp at: %s\n", ytdl)
	} else {
		fmt.Println(":: yt-dlp not found (will be bottled automatically on demand).")
	}
}

func main() {
	mem.TuneMemory()

	progName := filepath.Base(os.Args[0])
	args := os.Args[1:]

	restoreSession := false
	skipUpdate := false
	var playbackInputs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			printHelp(progName)
			return
		case "-v", "--version":
			fmt.Printf("Vibe-Fi v%s (Go, Bubble Tea, libmpv)\n", config.Version)
			return
		case "-b", "--bottle":
			printBottleStatus()
			return
		case "-r", "--restore":
			restoreSession = true
		case "--no-update":
			skipUpdate = true
		default:
			playbackInputs = append(playbackInputs, arg)
		}
	}

	// Ensure modern stream resolver (yt-dlp)
	_ = bottle.EnsureBottledYtdlp()

	// Initialize libmpv Audio Player
	mpvPlayer, err := player.NewMPVPlayer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: failed to initialize audio engine: %v\n", err)
		os.Exit(1)
	}
	defer mpvPlayer.Close()

	app := tui.NewAppModel(mpvPlayer)

	// Process initial CLI playback inputs
	var initialQueue []playlist.PlaylistSong
	startPlayback := false

	for _, input := range playbackInputs {
		fi, err := os.Stat(input)
		if err == nil && fi.IsDir() {
			// Scan directory recursively
			_ = filepath.WalkDir(input, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr == nil && !d.IsDir() && library.IsAudioFile(path) {
					absPath, _ := filepath.Abs(path)
					title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
					artist, _ := stringutil.CleanTrackTitle(title)
					initialQueue = append(initialQueue, playlist.PlaylistSong{Title: title, URL: absPath, Artist: artist})
				}
				return nil
			})
		} else if err == nil && !fi.IsDir() {
			// Single audio file
			absPath, _ := filepath.Abs(input)
			title := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
			artist, _ := stringutil.CleanTrackTitle(title)
			initialQueue = append(initialQueue, playlist.PlaylistSong{Title: title, URL: absPath, Artist: artist})
		} else if stringutil.IsURL(input) {
			if !net.IsOnline() {
				fmt.Fprintln(os.Stderr, ":: Error: Internet connection required to stream URL.")
				continue
			}
			fmt.Printf("Resolving stream: %s...\n", input)
			info, err := search.ResolveStreamInfo(input)
			if err == nil && info.StreamURL != "" {
				title := info.Title
				if info.Artist != "" && !strings.Contains(title, " - ") {
					title = info.Artist + " - " + title
				}
				if title == "" {
					title = input
				}
				initialQueue = append(initialQueue, playlist.PlaylistSong{
					Title:    title,
					URL:      info.StreamURL,
					Duration: stringutil.FormatDuration(info.Duration),
					Artist:   info.Artist,
				})
			}
		} else {
			// Search YouTube
			if !net.IsOnline() {
				fmt.Fprintln(os.Stderr, ":: Error: Internet connection required to search YouTube.")
				continue
			}
			fmt.Printf("Searching YouTube for: %s...\n", input)
			hits, err := search.SearchYouTube(input, 5)
			if err == nil && len(hits) > 0 {
				for _, hit := range hits {
					initialQueue = append(initialQueue, playlist.PlaylistSong{
						Title:    hit.Title,
						URL:      hit.URL,
						Duration: hit.Duration,
						Artist:   hit.Uploader,
					})
				}
			}
		}
	}

	if len(initialQueue) > 0 {
		app.PlayQueue = initialQueue
		app.QueueIndex = 0
		first := initialQueue[0]
		app.LastPlayedPath = first.URL
		app.CurrentLyricsTitle = first.Title
		_ = mpvPlayer.Load(first.URL, "replace")
		if first.Title != "" {
			_ = mpvPlayer.SetProperty("force-media-title", first.Title)
		}
		_ = mpvPlayer.Play()
		startPlayback = true
		app.SetMode(components.ViewModePlayback)
	}

	// Discord Rich Presence client (Unix domain socket IPC)
	discordClient := discord.NewClient("")
	if discordClient != nil {
		app.DiscordClient = discordClient
		defer discordClient.Close()
	}

	// Create Bubble Tea program
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Start D-Bus MPRIS server
	mprisServer := mpris.StartServer(mpvPlayer, p)
	if mprisServer != nil {
		app.MprisServer = mprisServer
		defer mprisServer.Stop()
	}

	// Wire EventBus subscribers for external integrations (MPRIS, Discord)
	if app.EventBus != nil {
		app.EventBus.Subscribe(eventbus.EventTrackStarted, func(e interface{}) {
			evt, ok := e.(eventbus.TrackStartedEvent)
			if !ok {
				return
			}
			if mprisServer != nil {
				mprisServer.EmitTrack(evt.Title, evt.Artist, evt.Duration)
				mprisServer.EmitPlaybackStatus("Playing")
			}
			if discordClient != nil {
				artist := evt.Artist
				trackName := evt.Title
				if artist == "" {
					artist, trackName = stringutil.CleanTrackTitle(evt.Title)
				}
				if artist == "" {
					artist = "Unknown Artist"
				}
				discordClient.UpdatePresence(trackName, artist)
			}
		})

		app.EventBus.Subscribe(eventbus.EventPlaybackStateChanged, func(e interface{}) {
			evt, ok := e.(eventbus.PlaybackStateChangedEvent)
			if !ok {
				return
			}
			if mprisServer != nil {
				mprisServer.EmitPlaybackStatus(evt.State.String())
			}
			if evt.State == eventbus.StateStopped && discordClient != nil {
				discordClient.ClearPresence()
			}
		})

		app.EventBus.Subscribe(eventbus.EventVolumeChanged, func(e interface{}) {
			evt, ok := e.(eventbus.VolumeChangedEvent)
			if !ok {
				return
			}
			if mprisServer != nil {
				mprisServer.EmitVolume(evt.Volume)
			}
		})

		app.EventBus.Subscribe(eventbus.EventSeeked, func(e interface{}) {
			evt, ok := e.(eventbus.SeekedEvent)
			if !ok {
				return
			}
			if mprisServer != nil {
				posMicro := int64(evt.PositionSeconds * 1000000.0)
				mprisServer.EmitSeeked(posMicro)
			}
		})
	}

	// If initial CLI track was supplied, broadcast domain event
	if startPlayback && len(initialQueue) > 0 {
		first := initialQueue[0]
		if app.EventBus != nil {
			app.EventBus.Publish(eventbus.EventTrackStarted, eventbus.TrackStartedEvent{
				Title:    first.Title,
				Artist:   first.Artist,
				URL:      first.URL,
				Duration: stringutil.ParseDuration(first.Duration),
			})
		}
	}

	// Background update check (non-blocking)
	if !skipUpdate {
		go func() {
			if ver, err := updater.CheckUpdate(); err == nil && ver != "" {
				p.Send(tui.UpdateDiscoveredMsg{Version: ver})
			}
		}()
	}

	// Graceful OS signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		p.Quit()
	}()

	if restoreSession {
		go func() {
			time.Sleep(100 * time.Millisecond)
			p.Send(tui.StatusMsg{Message: "Restoring session..."})
		}()
	} else if !startPlayback && len(playbackInputs) == 0 {
		app.SetMode(components.ViewModeIntro)
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
