package bottle

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"vibe-fi/internal/config"
)

// FindExecutable searches for an executable in bottle/bin first, then in PATH.
func FindExecutable(name string) string {
	candidates := []string{name}
	if runtime.GOOS == "windows" {
		candidates = []string{name + ".exe", name}
	}

	for _, cand := range candidates {
		bottleBin := filepath.Join(config.GetBottleBinDir(), cand)
		if fi, err := os.Stat(bottleBin); err == nil && !fi.IsDir() {
			if runtime.GOOS == "windows" || fi.Mode()&0111 != 0 {
				return bottleBin
			}
		}

		if p, err := exec.LookPath(cand); err == nil {
			return p
		}
	}

	return ""
}

// EnsureBottledYtdlp checks if yt-dlp is installed; if not, downloads the standalone release to ~/.vibe-fi/bottle/bin/yt-dlp.
func EnsureBottledYtdlp() error {
	if p := FindExecutable("yt-dlp"); p != "" {
		return nil
	}

	binDir := config.GetBottleBinDir()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	exeName := "yt-dlp"
	var downloadURL string
	switch runtime.GOOS {
	case "darwin":
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	case "windows":
		exeName = "yt-dlp.exe"
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	default:
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	}

	target := filepath.Join(binDir, exeName)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download yt-dlp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading yt-dlp: %s", resp.Status)
	}

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}
