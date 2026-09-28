package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"vibe-fi/internal/config"
	"vibe-fi/internal/utils/net"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckUpdate checks GitHub releases for a newer version (rate-limited to 24 hours).
func CheckUpdate() (string, error) {
	state, _ := config.LoadState()
	now := time.Now().Unix()

	// Rate limit check: only query GitHub once every 24 hours
	if state.LastUpdateCheck > 0 && (now-state.LastUpdateCheck) < 86400 {
		return state.AvailableUpdate, nil
	}

	if !net.IsOnline() {
		return "", fmt.Errorf("offline")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/Swadesh-c0de/vibe-fi/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "vibe-fi-go/"+config.Version)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}

	state.LastUpdateCheck = now
	latestTag := rel.TagName
	if latestTag != "" && latestTag != "v"+config.Version && latestTag != config.Version {
		state.AvailableUpdate = latestTag
		_ = config.SaveState(state)
		return latestTag, nil
	}

	state.AvailableUpdate = ""
	_ = config.SaveState(state)
	return "", nil
}
