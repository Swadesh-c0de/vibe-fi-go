package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SessionState stores persistent session settings from state.ini.
type SessionState struct {
	Path                   string
	Title                  string
	Position               float64
	Volume                 int
	Playlist               string
	Index                  int
	Theme                  string
	Visualizer             int
	Autoplay               bool
	AvailableUpdate        string
	LastUpdateCheck        int64
	UpdateDismissedVersion string
}

// DefaultState returns the default session configuration.
func DefaultState() SessionState {
	return SessionState{
		Volume:     100,
		Index:      -1,
		Theme:      DefaultTheme,
		Visualizer: DefaultVisualizer,
		Autoplay:   true,
	}
}

// LoadState reads ~/.vibe-fi/state.ini (or fallback ~/.vibe-fi-state.ini).
func LoadState() (SessionState, error) {
	state := DefaultState()

	stateFile := filepath.Join(GetVibeDir(), "state.ini")
	f, err := os.Open(stateFile)
	if err != nil {
		home, _ := os.UserHomeDir()
		if home != "" {
			f, err = os.Open(filepath.Join(home, ".vibe-fi-state.ini"))
		}
	}
	if err != nil {
		return state, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq == -1 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])

		switch key {
		case "path":
			state.Path = val
		case "title":
			state.Title = val
		case "position":
			if p, err := strconv.ParseFloat(val, 64); err == nil {
				state.Position = p
			}
		case "volume":
			if v, err := strconv.Atoi(val); err == nil {
				state.Volume = v
			}
		case "playlist":
			state.Playlist = val
		case "index":
			if idx, err := strconv.Atoi(val); err == nil {
				state.Index = idx
			}
		case "theme":
			state.Theme = val
		case "visualizer":
			if viz, err := strconv.Atoi(val); err == nil {
				state.Visualizer = viz
			}
		case "autoplay":
			state.Autoplay = (val == "1" || strings.ToLower(val) == "true")
		case "available_update":
			state.AvailableUpdate = val
		case "last_update_check":
			if t, err := strconv.ParseInt(val, 10, 64); err == nil {
				state.LastUpdateCheck = t
			}
		case "update_dismissed":
			state.UpdateDismissedVersion = val
		}
	}

	if err := scanner.Err(); err != nil {
		return state, err
	}

	return state, nil
}

// SaveState writes ~/.vibe-fi/state.ini in human-readable format.
func SaveState(state SessionState) error {
	stateFile := filepath.Join(GetVibeDir(), "state.ini")
	f, err := os.Create(stateFile)
	if err != nil {
		return err
	}
	defer f.Close()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("path=%s\n", state.Path))
	sb.WriteString(fmt.Sprintf("title=%s\n", state.Title))
	sb.WriteString(fmt.Sprintf("position=%.2f\n", state.Position))
	sb.WriteString(fmt.Sprintf("volume=%d\n", state.Volume))
	sb.WriteString(fmt.Sprintf("playlist=%s\n", state.Playlist))
	sb.WriteString(fmt.Sprintf("index=%d\n", state.Index))
	sb.WriteString(fmt.Sprintf("theme=%s\n", state.Theme))
	sb.WriteString(fmt.Sprintf("visualizer=%d\n", state.Visualizer))

	autoplayVal := "0"
	if state.Autoplay {
		autoplayVal = "1"
	}
	sb.WriteString(fmt.Sprintf("autoplay=%s\n", autoplayVal))

	if state.AvailableUpdate != "" {
		sb.WriteString(fmt.Sprintf("available_update=%s\n", state.AvailableUpdate))
	}
	if state.LastUpdateCheck > 0 {
		sb.WriteString(fmt.Sprintf("last_update_check=%d\n", state.LastUpdateCheck))
	}
	if state.UpdateDismissedVersion != "" {
		sb.WriteString(fmt.Sprintf("update_dismissed=%s\n", state.UpdateDismissedVersion))
	}

	_, err = f.WriteString(sb.String())
	return err
}
