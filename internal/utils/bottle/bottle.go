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
	bottleBin := filepath.Join(config.GetBottleBinDir(), name)
	if fi, err := os.Stat(bottleBin); err == nil && !fi.IsDir() && (fi.Mode()&0111 != 0) {
		return bottleBin
	}

	if p, err := exec.LookPath(name); err == nil {
		return p
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

	target := filepath.Join(binDir, "yt-dlp")

	var downloadURL string
	if runtime.GOOS == "darwin" {
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	} else {
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	}

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
